package main

import (
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Run the tests using the command go test
// and run the tests and the benchmarks with go test -bench=.

// Test setting and then getting the cache
func TestCacheRoundTrip(t *testing.T) {
	ip := net.ParseIP("192.168.1.1")
	setCachedIP(ip, IPv4)
	cache, err := getCachedIP(IPv4)
	if err != nil {
		t.Fatal(err)
	}
	if !cache.IPAddress.Equal(ip) {
		t.Errorf("Expected %s, got %s", ip, cache.IPAddress)
	}
	if cache.RecordType != "A" {
		t.Errorf("Expected A, got %s", cache.RecordType)
	}

	t.Log(cache)
}

// Make sure the timestamp is recent after a set
func TestCacheTimeIsRecent(t *testing.T) {
	setCachedIP(net.ParseIP("127.0.0.1"), IPv4)
	cache, err := getCachedIP(IPv4)
	if err != nil {
		t.Fatal(err)
	}
	if time.Since(cache.Time) > 2*time.Second {
		t.Errorf("Cache time is not recent: %s", cache.Time)
	}
}

// Test DisableCFCache
func TestDisableCFCachePreventsWrite(t *testing.T) {
	conf.DisableCFCache = true
	defer func() { conf.DisableCFCache = false }()

	// Remove any existing cache first
	os.Remove(getCacheFilePath("A"))
	setCachedIP(net.ParseIP("1.2.3.4"), IPv4)

	_, err := getCachedIP(IPv4)
	if err == nil {
		t.Error("Expected error when cache is disabled, but got none")
	}
}

// Test handling a missing file
func TestGetCachedIPMissingFile(t *testing.T) {
	os.Remove(getCacheFilePath(IPv6.getRecordType()))
	_, err := getCachedIP(IPv6)
	if err == nil {
		t.Error("Expected error for missing cache file, got nil")
	}
}

// Test handling a corrupted file
func TestGetCachedIPCorruptFile(t *testing.T) {
	path := getCacheFilePath(IPv4.getRecordType())
	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, []byte("not valid json{{{"), 0664)

	_, err := getCachedIP(IPv4)
	if err == nil {
		t.Error("Expected error for corrupt cache file, got nil")
	}
}
