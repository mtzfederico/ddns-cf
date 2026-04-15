package main

import (
	"net"
	"testing"
)

// Run the tests using the command go test
// and run the tests and the benchmarks with go test -bench=.

func TestSetCachedIP(t *testing.T) {
	setCachedIP(net.ParseIP("127.0.0.2"), IPv6)
}

func TestGetCachedIP(t *testing.T) {
	cache, err := getCachedIP(IPv6)
	if err != nil {
		t.Error(err)
	}

	t.Log(cache)
}
