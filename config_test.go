package main

import (
	"testing"
)

// Run the tests using the command go test
// and run the tests and the benchmarks with go test -bench=.

func TestParseConfig(t *testing.T) {
	var conf Config

	conf.get("sampleConfig.yaml")

	if conf.name != "<subdomain>.<domain.tld>" {
		t.Errorf("Unexpected Domain value, got: %s", conf.Domain)
	}

	if conf.Domain != "<domain.tld>" {
		t.Errorf("Unexpected Domain value, got: %s", conf.Domain)
	}

	if conf.DomainZoneID != "<DomainZoneID>" {
		t.Errorf("Unexpected DomainZoneID value, got: %s", conf.Domain)
	}

	if conf.SubDomainToUpdate != "<subdomain>" {
		t.Errorf("Unexpected SubDomainToUpdate value, got: %s", conf.Domain)
	}

	if conf.APIKey != "<Your API Key>" {
		t.Errorf("Unexpected APIKey value, got: %s", conf.Domain)
	}

	// Not set in sampleConfig. It should default to 0.
	if conf.RecordTTL != 0 {
		t.Errorf("Unexpected RecordTTL value, got: %s", conf.Domain)
	}

	if conf.IsProxied != false {
		t.Errorf("Unexpected IsProxied value, got: %s", conf.Domain)
	}

	if conf.DisableIPv4 != false {
		t.Errorf("Unexpected DisableIPv4 value, got: %s", conf.Domain)
	}

	if conf.DisableIPv6 != false {
		t.Errorf("Unexpected DisableIPv6 value, got: %s", conf.Domain)
	}

	if conf.ScriptOnChange != "myScript.sh" {
		t.Errorf("Unexpected ScriptOnChange value, got: %s", conf.Domain)
	}

	if conf.LogFile != "" {
		t.Errorf("Unexpected LogFile value, got: %s", conf.Domain)
	}

	if conf.LogLevel != "debug" {
		t.Errorf("Unexpected DebugLevel value, got: %s", conf.Domain)
	}
}
