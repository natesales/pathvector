package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/go-ping/ping"
	log "github.com/sirupsen/logrus"
)

var globalOptimizer Optimizer

// optimizationDelimiter is an arbitrary delimiter used to split ASN from peerName
var optimizationDelimiter = "####"

// probeResult stores a single probe result
type probeResult struct {
	Time  int64
	Stats ping.Statistics
}

type peerAvg struct {
	Latency    time.Duration
	PacketLoss float64
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
	pinger.SetNetwork("ip") // TODO: Is this needed?
	pinger.SetPrivileged(!udp)

	// Run the ping
	err = pinger.Run()
	if err != nil {
		return &ping.Statistics{}, err
	}

	return pinger.Statistics(), nil // nil error
}

// startProbe starts the probe scheduler to send probes to all configured targets and logs the results
func startProbe(sourceMap map[string][]string) error {
	// Initialize Db map
	if globalOptimizer.Db == nil {
		globalOptimizer.Db = map[string][]probeResult{} // peerName to list of probe results
	}

	for {
		// Loop over every source/target pair
		for peerName, sources := range sourceMap {
			for _, source := range sources {
				for _, target := range globalOptimizer.Targets {
					if sameAddressFamily(source, target) {
						log.Debugf("[Optimizer] Sending %d ICMP probes src %s dst %s", globalOptimizer.PingCount, source, target)
						stats, err := sendPing(source, target, globalOptimizer.PingCount, globalOptimizer.PingTimeout, probeUdpMode)
						if err != nil {
							return err
						}

						// Check for nil Db entries
						if globalOptimizer.Db[peerName] == nil {
							globalOptimizer.Db[peerName] = []probeResult{}
						}

						result := probeResult{
							Time:  time.Now().UnixNano(),
							Stats: *stats,
						}

						log.Debugf("[Optimizer] cache usage: %d/%d", len(globalOptimizer.Db[peerName]), globalOptimizer.CacheSize)

						if len(globalOptimizer.Db[peerName]) < globalOptimizer.CacheSize {
							// If the array is not full to CacheSize, append the result
							globalOptimizer.Db[peerName] = append(globalOptimizer.Db[peerName], result)
						} else {
							// If the array is full to probeCacheSize...
							if exitOnCacheFull {
								return nil
							} else {
								// Chop off the first element and append the result
								globalOptimizer.Db[peerName] = append(globalOptimizer.Db[peerName][1:], result)
							}
						}
					}
				}
			}
		}

		// Compute averages
		computeMetrics()

		// Sleep before sending the next probe
		waitInterval := time.Duration(globalOptimizer.Interval) * time.Second
		log.Debugf("[Optimizer] Waiting %s until next probe run", waitInterval)
		time.Sleep(waitInterval)
	}
}

// computeMetrics calculates average latency and packet loss
func computeMetrics() {
	p := map[string]*peerAvg{}
	for peer := range globalOptimizer.Db {
		if p[peer] == nil {
			p[peer] = &peerAvg{Latency: 0, PacketLoss: 0}
		}
		for result := range globalOptimizer.Db[peer] {
			p[peer].PacketLoss += globalOptimizer.Db[peer][result].Stats.PacketLoss
			p[peer].Latency += globalOptimizer.Db[peer][result].Stats.AvgRtt
		}

		// Calculate average latency and packet loss
		totalProbes := float64(len(globalOptimizer.Db[peer]))
		p[peer].PacketLoss = p[peer].PacketLoss / totalProbes
		p[peer].Latency = p[peer].Latency / time.Duration(totalProbes)

		// Check thresholds to apply optimizations
		var alerts []string
		peerASN, peerName := parsePeerDelimiter(peer)
		if p[peer].PacketLoss >= globalOptimizer.PacketLossThreshold {
			alerts = append(alerts, fmt.Sprintf("Peer AS%s %s exceeded maximum allowable packet loss: %f >= %f", peerASN, peerName, p[peer].PacketLoss, globalOptimizer.PacketLossThreshold))
		}
		if p[peer].Latency >= time.Duration(globalOptimizer.LatencyThreshold)*time.Millisecond {
			alerts = append(alerts, fmt.Sprintf("Peer AS%s %s exceeded maximum allowable latency: %v >= %v", peerASN, peerName, p[peer].Latency, globalOptimizer.LatencyThreshold))
		}

		// If there is at least one alert,
		if len(alerts) > 0 {
			for _, alert := range alerts {
				log.Debugf("[Optimizer] %s", alert)
				if globalOptimizer.AlertScript != "" {
					birdCmd := exec.Command(globalOptimizer.AlertScript, alert)
					birdCmd.Stdout = os.Stdout
					birdCmd.Stderr = os.Stderr
					if err := birdCmd.Run(); err != nil {
						log.Warnf("[Optimizer] alert script: %v", err)
					}
				}
			}
			optimizePeer(peer)
		}
	}
}

// parsePeerDelimiter parses a ASN/name string and returns the ASN and name
func parsePeerDelimiter(i string) (string, string) {
	parts := strings.Split(i, optimizationDelimiter)
	return parts[0], parts[1]
}

func optimizePeer(peer string) {
	peerASN, peerName := parsePeerDelimiter(peer)
	fileName := path.Join(cacheDirectory, fmt.Sprintf("AS%s_%s.conf", peerASN, *sanitize(peerName)))
	peerFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal("reading peer file: " + err.Error())
	}
	peerData := globalConfig.Peers[peerName]

	if *peerData.OptimizeInbound {
		// Calculate new local pref
		currentLocalPref := *peerData.LocalPref
		newLocalPref := uint(currentLocalPref) - globalOptimizer.LocalPrefModifier

		lpRegex := regexp.MustCompile(`bgp_local_pref = .*; # pathvector:localpref`)
		modified := lpRegex.ReplaceAllString(string(peerFile), fmt.Sprintf("bgp_local_pref = %d; # pathvector:localpref", newLocalPref))

		if err := ioutil.WriteFile(fileName, []byte(modified), 0755); err != nil {
			log.Fatal(err)
		} else {
			log.Printf("[Optimizer] Lowered AS%s %s local-pref from %d to %d", peerASN, peerName, currentLocalPref, newLocalPref)
		}
	}

	// Run BIRD config validation
	birdValidate()

	if !dryRun {
		moveCacheAndReconfig()
	}
}
