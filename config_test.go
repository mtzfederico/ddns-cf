package main

import (
	"testing"
)

// Run the tests using the command go test
// and run the tests and the benchmarks with go test -bench=.

func TestParseConfig(t *testing.T) {
	var Config conf

	Config.get("sampleConfig.yaml")

	if Config._Name != "<subdomain>.<domain.tld>" {
		t.Errorf("Unexpected Domain value, got: %s", Config.Domain)
	}

	if Config.Domain != "<domain.tld>" {
		t.Errorf("Unexpected Domain value, got: %s", Config.Domain)
	}

	if Config.DomainZoneID != "<DomainZoneID>" {
		t.Errorf("Unexpected DomainZoneID value, got: %s", Config.Domain)
	}

	if Config.SubDomainToUpdate != "<subdomain>" {
		t.Errorf("Unexpected SubDomainToUpdate value, got: %s", Config.Domain)
	}

	if Config.APIKey != "<Your API Key>" {
		t.Errorf("Unexpected APIKey value, got: %s", Config.Domain)
	}

	// Not set in sampleConfig. It should default to 0.
	if Config.RecordTTL != 0 {
		t.Errorf("Unexpected RecordTTL value, got: %s", Config.Domain)
	}

	if Config.IsProxied != false {
		t.Errorf("Unexpected IsProxied value, got: %s", Config.Domain)
	}

	if Config.DisableIPv4 != false {
		t.Errorf("Unexpected DisableIPv4 value, got: %s", Config.Domain)
	}

	if Config.DisableIPv6 != false {
		t.Errorf("Unexpected DisableIPv6 value, got: %s", Config.Domain)
	}

	if Config.ScriptOnChange != "myScript.sh" {
		t.Errorf("Unexpected ScriptOnChange value, got: %s", Config.Domain)
	}

	if Config.LogFile != "" {
		t.Errorf("Unexpected LogFile value, got: %s", Config.Domain)
	}

	if Config.LogLevel != "debug" {
		t.Errorf("Unexpected DebugLevel value, got: %s", Config.Domain)
	}
}
