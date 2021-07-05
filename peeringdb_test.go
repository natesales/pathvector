package main

import (
	"strings"
	"testing"
)

func TestPeeringDbQuery(t *testing.T) {
	testCases := []struct {
		asn          int
		asSet        string
		name         string
		importLimit4 int
		importLimit6 int
		shouldError  bool
	}{
		{112, "AS112", "DNS-OARC-112", 2, 2, false},
		{20144, "AS-LROOT", "l.root-servers.net", 5, 5, false},
		{25152, "RIPE::RS-KROOT RIPE::RS-KROOT-V6", "RIPE NCC K-Root Operations", 5, 5, false},
		{65530, "RIPE::RS-KROOT RIPE::RS-KROOT-V6", "RIPE NCC K-Root Operations", 5, 5, true}, // Private ASN, no PeeringDB page
	}
	for _, tc := range testCases {
		pDbData, err := getPeeringDbData(tc.asn)
		if err != nil && !tc.shouldError {
			t.Error(err)
		}

		if tc.shouldError && err == nil {
			t.Errorf("asn %d should have errored but didnt", tc.asn)
		}

		if tc.shouldError && err != nil && !strings.Contains(err.Error(), "doesn't have a PeeringDB page") {
			t.Errorf("asn %d should have thrown a no PeeringDB error but got a different error: %v", tc.asn, err)
		}

		if err == nil {
			if pDbData.ASSet != tc.asSet {
				t.Errorf("expected as-set %s got %s", tc.asSet, pDbData.ASSet)
			}
			if pDbData.Name != tc.name {
				t.Errorf("expected name %s got %s", tc.name, pDbData.Name)
			}
			if pDbData.ImportLimit4 != tc.importLimit4 {
				t.Errorf("expected IPv4 import limit %d got %d", tc.importLimit4, pDbData.ImportLimit4)
			}
			if pDbData.ImportLimit6 != tc.importLimit6 {
				t.Errorf("expected IPv6 import limit %d got %d", tc.importLimit6, pDbData.ImportLimit6)
			}
		}
	}
}

func TestPeeringDbNoPage(t *testing.T) {
	_, err := getPeeringDbData(65530)
	if err == nil || !strings.Contains(err.Error(), "doesn't have a PeeringDB page") {
		t.Errorf("expected PeeringDB page not exist error, got %v", err)
	}
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func TestPeeringDbQueryAndModify(t *testing.T) {
	testCases := []struct {
		asn         int
		auto        bool
		shouldError bool
	}{
		{112, true, false},
		{112, false, false},
	}
	for _, tc := range testCases {
		err := runPeeringDbQuery(&peer{
			ASN:              intPtr(tc.asn),
			AutoImportLimits: boolPtr(tc.auto),
			AutoASSet:        boolPtr(tc.auto),
			ImportLimit4:     intPtr(0),
			ImportLimit6:     intPtr(0),
		})

		// If it errored but shouldn't have
		if err != nil && !tc.shouldError {
			t.Error(err)
		}

		// If it didn't error but should have
		if tc.shouldError && err == nil {
			t.Errorf("asn %d should have errored but didnt", tc.asn)
		}
	}
}

func TestSanitizeASSet(t *testing.T) {
	testCases := []struct {
		asSet    string
		expected string
	}{
		{"AS34553:AS-ALL", "AS34553:AS-ALL"},
		{"RIPE::AS34553:AS-ALL", "AS34553:AS-ALL"},
		{"RADB::AS-HURRICANE RADB::AS-HURRICANEV6", "AS-HURRICANE"},
	}
	for _, tc := range testCases {
		out := sanitizeASSet(tc.asSet)
		if out != tc.expected {
			t.Errorf("expected %s got %s", tc.expected, out)
		}
	}
}
