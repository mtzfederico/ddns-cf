package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
)

// A struct used to cached the IP address stored in Cloudlare
type IPCache struct {
	// The IP address cached
	IPAddress string `json:"IPAddress" binding:"required"`
	// The type of DNS record cached (A or AAAA).
	RecordType string `json:"RecordType" binding:"required"`
	// The time that it was last checked
	Time time.Time `json:"Time" binding:"required"`
}

func getCachedIP(version IPVersion) (IPCache, error) {
	recordType := version.getRecordType()

	path := filepath.Join(os.TempDir(), "ddns-cf-"+recordType+"-cache.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return IPCache{IPAddress: "", RecordType: IPv6.getRecordType(), Time: time.UnixMicro(1)}, err
	}

	var cache IPCache
	err = json.Unmarshal(data, &cache)
	if err != nil {
		return IPCache{IPAddress: "", RecordType: IPv6.getRecordType(), Time: time.UnixMicro(1)}, err
	}

	return cache, nil
}

func setCachedIP(address string, version IPVersion) {
	if conf.DisableCFCache {
		return
	}

	recordType := version.getRecordType()

	cache := IPCache{IPAddress: address, RecordType: recordType, Time: time.Now()}

	jsonData, err := json.Marshal(cache)
	if err != nil {
		log.WithFields(log.Fields{"err": err, "address": address, "RecordType": recordType}).Error("[setCachedIP] Failed to encode JSON")
		return
	}

	path := filepath.Join(os.TempDir(), "ddns-cf-"+recordType+"-cache.json")
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0664)

	if err != nil {
		log.WithFields(log.Fields{"err": err, "path": path, "address": address, "RecordType": recordType}).Error("[setCachedIP] Failed to create file")
		return
	}
	defer file.Close()

	_, err = file.WriteString(string(jsonData))
	if err != nil {
		log.WithFields(log.Fields{"err": err, "address": address, "RecordType": recordType}).Error("[setCachedIP] Failed to save file")
	}

	log.WithFields(log.Fields{"path": path}).Debug("[setCachedIP] Cache Set")
}
