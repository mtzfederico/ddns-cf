package main

import (
	"testing"
)

func TestGetRecordType(t *testing.T) {
	if IPv4.getRecordType() != "A" {
		t.Errorf("Expected A, got %s", IPv4.getRecordType())
	}
	if IPv6.getRecordType() != "AAAA" {
		t.Errorf("Expected AAAA, got %s", IPv6.getRecordType())
	}
	if IPVersion("invalid").getRecordType() != "" {
		t.Error("Expected empty string for invalid version")
	}
}
