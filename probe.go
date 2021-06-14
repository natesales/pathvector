package main

import (
	"time"

	"github.com/go-ping/ping"
)

type Probe struct {
	Sources []string // List of source addresses to send ICMP probes from
	Targets []string // List of target addresses to send ICMP probes to

	PingCount   int // Number of ICMP messages to send per probe
	PingTimeout int // Number of seconds to wait before considering the ICMP message unanswered
	Sleep       int // Number of seconds to wait between sending ICMP messages
	CacheSize   int // Number of results to store per source/target pair. There will be a total of probeCacheSize*len(sources)*len(targets) results stored.

	Db map[string]map[string][]probeResult // Probe ping result database
}

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

// Start starts the probe scheduler to send probes to all configured targets and logs the results
func (p *Probe) Start() error {
	// Initialize Db map
	if p.Db == nil {
		p.Db = map[string]map[string][]probeResult{}
	}

	// Loop over every source/target pair
	for {
		for _, source := range p.Sources {
			for _, target := range p.Targets {
				stats, err := sendPing(source, target, p.PingCount, p.PingTimeout)
				if err != nil {
					return err
				}

				// Check for nil Db entries
				if p.Db[source] == nil {
					p.Db[source] = map[string][]probeResult{}
				}
				if p.Db[source][target] == nil {
					p.Db[source][target] = []probeResult{}
				}

				result := probeResult{
					Time:  time.Now().UnixNano(),
					Stats: *stats,
				}

				// If the array is not full to probeCacheSize, append the result
				if len(p.Db[source][target]) < p.CacheSize {
					p.Db[source][target] = append(p.Db[source][target], result)
				} else {
					// If the array is full to probeCacheSize, chop off the first element and append the last
					p.Db[source][target] = append(p.Db[source][target][1:], result)
				}
			}
		}

		// Sleep before sending the next probe
		time.Sleep(time.Duration(p.Sleep) * time.Second)
	}
}

// AcquisitionProgress returns the percent value of how full the probe database is. A value of 1 represents completely full and ready to optimize.
func (p *Probe) AcquisitionProgress() float32 {
	totalEntries := p.CacheSize * len(p.Sources) * len(p.Targets) // Expected total number of entries to make up a 100% probe acquisition

	actualEntries := 0
	// For all source/target combinations, increment actualEntries by the number of entries recorded
	for source := range p.Db {
		for _, entries := range p.Db[source] {
			actualEntries += len(entries)
		}
	}

	// Return the fractional value
	return float32(actualEntries) / float32(totalEntries)
}
