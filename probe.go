package main

import (
	"time"

	"github.com/go-ping/ping"
)

// probeResult stores a single probe result
type probeResult struct {
	Time  int64
	Stats ping.Statistics
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
	pinger.SetPrivileged(true)

	// Run the ping
	err = pinger.Run()
	if err != nil {
		return &ping.Statistics{}, err
	}

	return pinger.Statistics(), nil // nil error
}

// startProbe starts the probe scheduler to send probes to all configured targets and logs the results
func startProbe(o optimizer, sourceMap map[string][]string) error {
	// Initialize Db map
	if o.Db == nil {
		o.Db = map[string][]probeResult{} // peerName to list of probe results
	}

	// Loop over every source/target pair
	for {
		for peerName, sources := range sourceMap {
			for _, source := range sources {
				for _, target := range o.Targets {
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

					// If the array is not full to probeCacheSize, append the result
					if len(o.Db[peerName]) < o.CacheSize {
						o.Db[peerName] = append(o.Db[target], result)
					} else {
						// If the array is full to probeCacheSize, chop off the first element and append the last
						o.Db[peerName] = append(o.Db[peerName][1:], result)
					}
				}
			}
		}

		// Sleep before sending the next probe
		time.Sleep(time.Duration(o.Interval) * time.Second)
	}
}

// AcquisitionProgress returns the percent value of how full the probe database is. A value of 1 represents completely full and ready to optimize.
//func (p *optimizer) AcquisitionProgress() float32 {
//	totalEntries := p.CacheSize * len(p.Sources) * len(p.Targets) // Expected total number of entries to make up a 100% probe acquisition
//
//	actualEntries := 0
//	// For all source/target combinations, increment actualEntries by the number of entries recorded
//	for source := range p.Db {
//		for _, entries := range p.Db[source] {
//			actualEntries += len(entries)
//		}
//	}
//
//	// Return the fractional value
//	return float32(actualEntries) / float32(totalEntries)
//}
