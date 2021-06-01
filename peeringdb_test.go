package main

import (
	"strings"
	"testing"
)

func TestPeeringDbQuery(t *testing.T) {
	pDbData, err := getPeeringDbData(112)
	if err != nil {
		t.Error(err)
	}
	if pDbData.ASSet != "AS112" {
		t.Errorf("expected as-set AS112 got %s", pDbData.ASSet)
	}
	if pDbData.Name != "DNS-OARC-112" {
		t.Errorf("expected name DNS-OARC-112 got %s", pDbData.Name)
	}
	if pDbData.ImportLimit4 != 2 {
		t.Errorf("expected IPv4 import limit 2 got %d", pDbData.ImportLimit4)
	}
	if pDbData.ImportLimit6 != 2 {
		t.Errorf("expected IPv6 import limit 2 got %d", pDbData.ImportLimit6)
	}
}

func TestPeeringDbNoPage(t *testing.T) {
	_, err := getPeeringDbData(65530)
	if err == nil || !strings.Contains(err.Error(), "doesn't have a PeeringDB page") {
		t.Errorf("expected PeeringDB page not exist error, got %v", err)
	}
}
