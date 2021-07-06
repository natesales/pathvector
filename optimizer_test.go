package main

import (
	"github.com/go-ping/ping"
	"testing"
)

func TestOptimizerSameAddressFamily(t *testing.T) {
	testCases := []struct {
		a    string
		b    string
		same bool
	}{
		{"192.0.2.1", "192.0.2.1", true},
		{"192.0.2.1", "2001:db8::1", false},
		{"2001:db8::1", "2001:db8::1", true},
		{"2001:db8::1", "192.0.2.1", false},
	}
	for _, tc := range testCases {
		out := sameAddressFamily(tc.a, tc.b)
		if out != tc.same {
			t.Errorf("a %s b %s expected same %v got %v", tc.a, tc.b, tc.same, out)
		}
	}
}

func TestOptimizerAcquisitionProgress(t *testing.T) {
	testCases := []struct {
		optimizer Optimizer
		numPeers  int
		expected  float64
	}{
		{Optimizer{CacheSize: 5}, 1, 0}, // Empty cache
		{Optimizer{CacheSize: 5, Db: map[string][]probeResult{"Example": {
			probeResult{Time: 0, Stats: ping.Statistics{}},
			probeResult{Time: 1, Stats: ping.Statistics{}},
			probeResult{Time: 2, Stats: ping.Statistics{}},
			probeResult{Time: 3, Stats: ping.Statistics{}},
			probeResult{Time: 4, Stats: ping.Statistics{}},
		}}}, 1, 1}, // Full cache
		{Optimizer{CacheSize: 5, Db: map[string][]probeResult{"Example": {
			probeResult{Time: 0, Stats: ping.Statistics{}},
			probeResult{Time: 1, Stats: ping.Statistics{}},
		}}}, 1, 0.4}, // Partially full cache
	}
	for _, tc := range testCases {
		out := acquisitionProgress(tc.optimizer, tc.numPeers)
		if out != tc.expected {
			t.Errorf("expected same %f got %f", tc.expected, out)
		}
	}
}
