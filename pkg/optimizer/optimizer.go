package optimizer

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/go-ping/ping"
	log "github.com/sirupsen/logrus"

	"github.com/natesales/pathvector/pkg/bird"
	"github.com/natesales/pathvector/pkg/config"
	"github.com/natesales/pathvector/pkg/util"
)

// Delimiter is an arbitrary delimiter used to split ASN from peerName
var Delimiter = "####"

type peerAvg struct {
	Latency    time.Duration
	PacketLoss float64
}

// parsePeerDelimiter parses a ASN/name string and returns the ASN and name
func parsePeerDelimiter(i string) (string, string) {
	parts := strings.Split(i, Delimiter)
	return parts[0], parts[1]
}

// sameAddressFamily returns if two strings (IP addresses) are of the same address family
func sameAddressFamily(a string, b string) bool {
	a4 := net.ParseIP(a).To4() != nil // Is address A IPv4?
	b4 := net.ParseIP(b).To4() != nil // Is address B IPv4?
	// Are (both A and B IPv4) or (both A and B not IPv4)
	return (a4 && b4) || (!a4 && !b4)
}

// sendPing sends a probe ping to a specified target
func sendPing(source string, target string, count int, timeout int, udp bool) (*ping.Statistics, error) {
	pinger, err := ping.NewPinger(target)
	if err != nil {
		return &ping.Statistics{}, err
	}

	// Set pinger options
	pinger.Count = count
	pinger.Timeout = time.Duration(timeout) * time.Second
	pinger.Source = source
	pinger.SetPrivileged(!udp)

	// Run the ping
	if err = pinger.Run(); err != nil {
		return &ping.Statistics{}, fmt.Errorf("ping: %s", err)
	}

	return pinger.Statistics(), nil // nil error
}

// StartProbe starts the probe scheduler to send probes to all configured targets and logs the results
func StartProbe(o *config.Optimizer, sourceMap map[string][]string, global *config.Config, noConfigure bool, dryRun bool) error {
	// Initialize Db map
	if o.Db == nil {
		o.Db = map[string][]config.ProbeResult{} // peerName to list of probe results
	}

	for {
		// Loop over every source/target pair
		for peerName, sources := range sourceMap {
			for _, source := range sources {
				for _, target := range o.Targets {
					if sameAddressFamily(source, target) {
						log.Debugf("[Optimizer] Sending %d ICMP probes src %s dst %s", o.PingCount, source, target)
						stats, err := sendPing(source, target, o.PingCount, o.PingTimeout, o.ProbeUDPMode)
						if err != nil {
							return err
						}

						// Check for nil Db entries
						if o.Db[peerName] == nil {
							o.Db[peerName] = []config.ProbeResult{}
						}

						result := config.ProbeResult{
							Time:  time.Now().UnixNano(),
							Stats: *stats,
						}

						log.Debugf("[Optimizer] cache usage: %d/%d", len(o.Db[peerName]), o.CacheSize)

						if len(o.Db[peerName]) < o.CacheSize {
							// If the array is not full to CacheSize, append the result
							o.Db[peerName] = append(o.Db[peerName], result)
						} else {
							// If the array is full to probeCacheSize...
							if o.ExitOnCacheFull {
								return nil
							}
							// Chop off the first element and append the result
							o.Db[peerName] = append(o.Db[peerName][1:], result)
						}
					}
				}
			}
		}

		// Compute averages
		computeMetrics(o, global, noConfigure, dryRun)

		// Sleep before sending the next probe
		waitInterval := time.Duration(o.Interval) * time.Second
		log.Debugf("[Optimizer] Waiting %s until next probe run", waitInterval)
		time.Sleep(waitInterval)
	}
}

// computeMetrics calculates average latency and packet loss
func computeMetrics(o *config.Optimizer, global *config.Config, noConfigure bool, dryRun bool) {
	p := map[string]*peerAvg{}
	for peer := range o.Db {
		if p[peer] == nil {
			p[peer] = &peerAvg{Latency: 0, PacketLoss: 0}
		}
		for result := range o.Db[peer] {
			p[peer].PacketLoss += o.Db[peer][result].Stats.PacketLoss
			p[peer].Latency += o.Db[peer][result].Stats.AvgRtt
		}

		// Calculate average latency and packet loss
		totalProbes := float64(len(o.Db[peer]))
		p[peer].PacketLoss = p[peer].PacketLoss / totalProbes
		p[peer].Latency = p[peer].Latency / time.Duration(totalProbes)

		// Check thresholds to apply optimizations
		var alerts []string
		peerASN, peerName := parsePeerDelimiter(peer)
		if p[peer].PacketLoss >= o.PacketLossThreshold {
			alerts = append(alerts, fmt.Sprintf("Peer AS%s %s met or exceeded maximum allowable packet loss: %f >= %f",
				peerASN, peerName, p[peer].PacketLoss, o.PacketLossThreshold))
		}
		if p[peer].Latency >= time.Duration(o.LatencyThreshold)*time.Millisecond {
			alerts = append(alerts, fmt.Sprintf("Peer AS%s %s met or exceeded maximum allowable latency: %v >= %v",
				peerASN, peerName, p[peer].Latency, o.LatencyThreshold))
		}

		// If there is at least one alert,
		if len(alerts) > 0 {
			for _, alert := range alerts {
				log.Debugf("[Optimizer] %s", alert)
				if o.AlertScript != "" {
					//nolint:golint,gosec
					birdCmd := exec.Command(o.AlertScript, alert)
					birdCmd.Stdout = os.Stdout
					birdCmd.Stderr = os.Stderr
					if err := birdCmd.Run(); err != nil {
						log.Warnf("[Optimizer] alert script: %v", err)
					}
				}
			}
			modifyPref(peer,
				global.Peers,
				o.LocalPrefModifier,
				global.CacheDirectory,
				global.BIRDDirectory,
				global.BIRDSocket,
				global.BIRDBinary,
				noConfigure,
				dryRun,
			)
		}
	}
}

func modifyPref(
	peerPair string,
	peers map[string]*config.Peer,
	localPrefModifier uint,
	cacheDirectory string,
	birdDirectory string,
	birdSocket string,
	birdBinary string,
	noConfigure bool,
	dryRun bool,
) {
	peerASN, peerName := parsePeerDelimiter(peerPair)
	fileName := path.Join(cacheDirectory, fmt.Sprintf("AS%s_%s.conf", peerASN, *util.Sanitize(peerName)))
	peerFile, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatalf("reading peer file: %s", err)
	}

	peerData := peers[peerName]
	if *peerData.OptimizeInbound {
		// Calculate new local pref
		currentLocalPref := *peerData.LocalPref
		newLocalPref := uint(currentLocalPref) - localPrefModifier

		lpRegex := regexp.MustCompile(`bgp_local_pref = .*; # pathvector:localpref`)
		modified := lpRegex.ReplaceAllString(string(peerFile), fmt.Sprintf("bgp_local_pref = %d; # pathvector:localpref", newLocalPref))

		//nolint:golint,gosec
		if err := os.WriteFile(fileName, []byte(modified), 0755); err != nil {
			log.Fatal(err)
		} else {
			log.Printf("[Optimizer] Lowered AS%s %s local-pref from %d to %d", peerASN, peerName, currentLocalPref, newLocalPref)
		}
	}

	// Run BIRD config validation
	bird.Validate(birdBinary, cacheDirectory)

	if !dryRun {
		bird.MoveCacheAndReconfigure(birdDirectory, cacheDirectory, birdSocket, noConfigure)
	}
}
