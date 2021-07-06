package main

import (
	"net"
	"time"

	"github.com/go-ping/ping"
	log "github.com/sirupsen/logrus"
)

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
func sendPing(source string, target string, count int, timeout int) (*ping.Statistics, error) {
	pinger, err := ping.NewPinger(target)
	if err != nil {
		return &ping.Statistics{}, err
	}

	// Set pinger options
	pinger.Count = count
	pinger.Timeout = time.Duration(timeout) * time.Second
	pinger.Source = source
	pinger.SetNetwork("ip") // TODO: Is this needed?
	pinger.SetPrivileged(true)

	// Run the ping
	err = pinger.Run()
	if err != nil {
		return &ping.Statistics{}, err
	}

	return pinger.Statistics(), nil // nil error
}

// startProbe starts the probe scheduler to send probes to all configured targets and logs the results
func startProbe(o Optimizer, sourceMap map[string][]string) error {
	// Initialize Db map
	if o.Db == nil {
		o.Db = map[string][]probeResult{} // peerName to list of probe results
	}

	// Loop over every source/target pair
	for {
		for peerName, sources := range sourceMap {
			for _, source := range sources {
				for _, target := range o.Targets {
					if sameAddressFamily(source, target) {
						log.Debugf("[Optimizer] Sending %d ICMP probes src %s dst %s", o.PingCount, source, target)
						stats, err := sendPing(source, target, o.PingCount, o.PingTimeout)
						if err != nil {
							return err
						}

						// Check for nil Db entries
						if o.Db[peerName] == nil {
							o.Db[peerName] = []probeResult{}
						}

						result := probeResult{
							Time:  time.Now().UnixNano(),
							Stats: *stats,
						}

						if len(o.Db[peerName]) < o.CacheSize {
							// If the array is not full to CacheSize, append the result
							o.Db[peerName] = append(o.Db[peerName], result)
						} else {
							// If the array is full to probeCacheSize, chop off the first element and append the result
							o.Db[peerName] = append(o.Db[peerName][1:], result)
						}
					}
				}
			}
		}

		// Only start optimizing if enough metrics have been acquired
		if acquisitionProgress(o, len(sourceMap)) >= o.AcquisitionThreshold {

		}

		// Sleep before sending the next probe
		waitInterval := time.Duration(o.Interval) * time.Second
		log.Debugf("[Optimizer] Waiting %s until next probe run", waitInterval)
		time.Sleep(waitInterval)
	}
}

// acquisitionProgress returns the percent value of how full the probe database is. A value of 1 represents completely full and ready to optimize.
func acquisitionProgress(o Optimizer, numPeers int) float64 {
	totalEntries := o.CacheSize * numPeers // Expected total number of entries to make up a 100% probe acquisition

	actualEntries := 0
	// For each peer, increment actualEntries by the number of entries recorded
	for peer := range o.Db { // For each peer in the database
		actualEntries += len(o.Db[peer])
	}

	percent := float64(actualEntries) / float64(totalEntries)
	log.Debugf("[Optimizer] Acquisition progress: %d/%d (%v%%)", actualEntries, totalEntries, percent*10)
	return percent
}

// computeMetrics calculates average latency and packet loss
func computeMetrics(o Optimizer) {
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
		// TODO: Email/hook script alerts
		if p[peer].PacketLoss >= o.PacketLossThreshold {
			log.Debugf("[Optimizer] Peer %s exceeded maximum allowable packet loss: %f >= %f", peer, p[peer].PacketLoss, o.PacketLossThreshold)
		}
		if p[peer].Latency >= time.Duration(o.LatencyThreshold)*time.Millisecond {
			log.Debugf("[Optimizer] Peer %s exceeded maximum allowable latency: %v >= %v", peer, p[peer].Latency, o.LatencyThreshold)
		}
	}
}
