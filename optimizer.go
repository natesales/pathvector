package main

import (
	"fmt"
	"io/ioutil"
	"net"
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
		// Break optimization loop (used for testing)
		if globalOptimizer.Disable {
			return nil
		}

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

						if len(globalOptimizer.Db[peerName]) < globalOptimizer.CacheSize {
							// If the array is not full to CacheSize, append the result
							globalOptimizer.Db[peerName] = append(globalOptimizer.Db[peerName], result)
						} else {
							// If the array is full to probeCacheSize, chop off the first element and append the result
							globalOptimizer.Db[peerName] = append(globalOptimizer.Db[peerName][1:], result)
						}
					}
				}
			}
		}

		// Only start optimizing if enough metrics have been acquired
		if acquisitionProgress(len(sourceMap)) >= globalOptimizer.AcquisitionThreshold {
			computeMetrics()
		}

		// Sleep before sending the next probe
		waitInterval := time.Duration(globalOptimizer.Interval) * time.Second
		log.Debugf("[Optimizer] Waiting %s until next probe run", waitInterval)
		time.Sleep(waitInterval)
	}
}

// acquisitionProgress returns the percent value of how full the probe database is. A value of 1 represents completely full and ready to optimize.
func acquisitionProgress(numPeers int) float64 {
	totalEntries := globalOptimizer.CacheSize * numPeers // Expected total number of entries to make up a 100% probe acquisition

	actualEntries := 0
	// For each peer, increment actualEntries by the number of entries recorded
	for peer := range globalOptimizer.Db { // For each peer in the database
		actualEntries += len(globalOptimizer.Db[peer])
	}

	percent := float64(actualEntries) / float64(totalEntries)
	log.Debugf("[Optimizer] Acquisition progress: %d/%d (%v%%)", actualEntries, totalEntries, percent*10)
	return percent
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
		// TODO: Email/hook script alerts
		if p[peer].PacketLoss >= globalOptimizer.PacketLossThreshold {
			log.Debugf("[Optimizer] Peer %s exceeded maximum allowable packet loss: %f >= %f", peer, p[peer].PacketLoss, globalOptimizer.PacketLossThreshold)
			optimizePeer(peer)
		}
		if p[peer].Latency >= time.Duration(globalOptimizer.LatencyThreshold)*time.Millisecond {
			log.Debugf("[Optimizer] Peer %s exceeded maximum allowable latency: %v >= %v", peer, p[peer].Latency, globalOptimizer.LatencyThreshold)
			optimizePeer(peer)
		}
	}
}

func optimizePeer(peer string) {
	s := strings.Split(peer, optimizationDelimiter)
	fileName := path.Join(cacheDirectory, fmt.Sprintf("AS%s_%s.conf", s[0], *sanitize(s[1])))
	peerFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal("reading peer file: " + err.Error())
	}
	peerData := globalConfig.Peers[s[1]]

	if *peerData.OptimizeInbound {
		// Calculate new local pref
		currentLocalPref := *peerData.LocalPref
		newLocalPref := uint(currentLocalPref) - globalOptimizer.LocalPrefModifier

		lpRegex := regexp.MustCompile(`bgp_local_pref = .*; # pathvector:localpref`)
		modified := lpRegex.ReplaceAllString(string(peerFile), fmt.Sprintf("bgp_local_pref = %d; # pathvector:localpref", newLocalPref))

		if err := ioutil.WriteFile(fileName, []byte(modified), 0755); err != nil {
			log.Fatal(err)
		} else {
			log.Printf("[Optimizer] Lowered %s local-pref from %d to %d", s[1], currentLocalPref, newLocalPref)
		}
	}

	// Run BIRD config validation
	birdValidate()

	if !dryRun {
		moveCacheAndReconfig()
	}
}
