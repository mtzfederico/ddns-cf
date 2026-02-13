package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Jeffail/gabs"
	log "github.com/sirupsen/logrus"
)

const (
	baseURL = "https://api.cloudflare.com/client/v4/"
)

var Config conf
var httpClient *http.Client

type RecordData struct {
	Type    string `json:"type" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Content string `json:"content" binding:"required"`
	TTL     int    `json:"ttl" binding:"required"`
	Proxied bool   `json:"proxied" binding:"required"`
}

func sendRequest(path string, method string, requestBody []byte) *gabs.Container {
	url := fmt.Sprintf("%s%s", baseURL, path)
	// fmt.Printf("%s%s %s%s\n", color.Yellow, method, url, color.Reset)
	log.WithFields(log.Fields{"method": method, "url": url}).Trace(("[sendRequest] Sending request"))

	var req *http.Request
	var err error
	if requestBody != nil {
		requestData := bytes.NewBuffer(requestBody)
		req, err = http.NewRequest(method, url, requestData)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		log.Fatal("Error creating Request: ", err)
	}

	cfAuthHeader := fmt.Sprintf("Bearer %s", Config.APIKey)
	req.Header.Set("Authorization", cfAuthHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ddns-cf/1.1 (github.com/mtzfederico/ddns-cf)")

	resp, err := httpClient.Do(req)

	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("[sendRequest] httpClient error")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	// fmt.Printf("%s%s%s", color.Ize(color.Blue, "----- Response Starts -----\n"), string(body), color.Ize(color.Blue, "\n----- Response Ends -----\n"))
	log.WithFields(log.Fields{"responseBody": body}).Trace(("[sendRequest] Received response"))

	jsonParsed, err := gabs.ParseJSON(body)

	if err != nil {
		log.WithFields(log.Fields{"path": path, "method": method, "responseBody": body}).Fatal(("[sendRequest] Failed to parse JSON"))
	}

	return jsonParsed
}

func getIP(ipVersion string) string {
	url := fmt.Sprintf("https://ip%s.icanhazip.com", ipVersion)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "ipVersion": ipVersion}).Fatal("[getIP] Error creating request")
	}

	req.Header.Set("User-Agent", "ddns-cf/1.0 (github.com/mtzfederico/ddns-cf)")

	resp, err := httpClient.Do(req)

	if err != nil {
		log.WithFields(log.Fields{"error": err, "ipVersion": ipVersion}).Fatal("[getIP] Error sending request")
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "resp": resp}).Fatal("[getIP] Error processing response")
	}
	return strings.TrimSuffix(string(body), "\n")
}

func getZoneID() string {
	// Get domain's zone id. data.result[0].id
	// https://api.cloudflare.com/#zone-list-zones
	url := fmt.Sprintf("zones?name=%s", Config.Domain)
	resp := sendRequest(url, "GET", nil)
	zoneID, ok := resp.S("result").Index(0).Path("id").Data().(string)
	if !ok {
		log.WithFields(log.Fields{"resp": resp}).Error("[getZoneID] Error decoding zoneID")
	}
	log.WithFields(log.Fields{"zoneID": zoneID}).Debug("[getZoneID] Got zoneID from CF")
	return zoneID
}

// Get the domain's current value for the specified record type (A, AAAA, TXT, etc.)
func getCurrentValue(recordType string) (string, string, error) {
	// https://api.cloudflare.com/#dns-records-for-a-zone-list-dns-records
	// name is the FQDN. 'subdomain.domain.tld' or 'domain.tld'
	zoneID := Config.DomainZoneID

	if zoneID == "" {
		log.Info("ZoneID not in config file, fetching from CF.")
		zoneID = getZoneID()
		Config.DomainZoneID = zoneID // save for later use but don't save to file
	}

	path := fmt.Sprintf("zones/%s/dns_records?type=%s&name=%s", zoneID, recordType, Config._Name)
	resp := sendRequest(path, "GET", nil)

	success, ok := resp.Path("success").Data().(bool)

	if !ok {
		return "", "", errors.New("failed to decode JSON")
	}

	if !success {
		log.WithFields(log.Fields{"resp": resp}).Error("[getCurrentValue] API call failed")
		error := resp.S("errors").Index(0)
		errorCode := error.Path("code").Data().(string)
		if errorCode == "7003" {
			return "", "", errors.New("domain/subdomain does not exist")
		}
		return "", "", fmt.Errorf("errorCode: %s: %s", errorCode, error.Path("message").Data().(string))
	}

	result := resp.S("result")
	resultLen, err := result.ArrayCount()

	if err != nil {
		return "", "", err
	}

	// the subdomain exists but there is no record for this type. There is an A record but no AAAA record or vice versa.
	if resultLen == 0 {
		return "", "", fmt.Errorf("no record of type %s for %s", recordType, Config._Name)
	}

	content, ok := result.Index(0).Path("content").Data().(string) // The record's value
	if !ok {
		return "", "", errors.New("failed to decode the record's value from JSON")
	}

	if content == "" {
		return content, "", fmt.Errorf("no Content for %s's %s record", Config._Name, recordType)

	}

	RecordID, ok := result.Index(0).Path("id").Data().(string) // The id of the actual A or AAAA record, needed to update it.
	if !ok {
		return "", "", errors.New("failed to decode the record's ID from JSON")
	}

	if RecordID == "" {
		return content, "", fmt.Errorf("no recordID for %s's %s record", Config._Name, recordType)
	}

	return content, RecordID, nil
}

func updateRecord(recordID string, recordType string, IP string) {
	// https://api.cloudflare.com/#dns-records-for-a-zone-update-dns-record
	path := fmt.Sprintf("zones/%s/dns_records/%s", Config.DomainZoneID, recordID)
	ttl := Config.RecordTTL
	if ttl == 0 {
		ttl = 1 // 1 is Automatic
	}

	var requestBody RecordData
	requestBody.Type = recordType
	requestBody.Name = Config._Name
	requestBody.Content = IP
	requestBody.TTL = ttl
	requestBody.Proxied = Config.IsProxied

	requestData, _ := json.Marshal((requestBody))
	resp := sendRequest(path, "PUT", requestData)

	success, ok := resp.S("success").Data().(bool)
	if !ok {
		log.WithFields(log.Fields{"resp": resp}).Error("[updateRecord] Error decoding response")
	}

	if !success {
		errorMessage, _ := resp.S("errors").Index(0).Path("message").Data().(string)
		log.WithFields(log.Fields{"errorMessage": errorMessage}).Error("[updateRecord] Failed to update the record")
		return
	}
	log.WithFields(log.Fields{"recordType": recordType}).Info("record changed successfully")
}

func createRecord(recordType string, IP string) {
	// https://api.cloudflare.com/#dns-records-for-a-zone-create-dns-record
	path := fmt.Sprintf("zones/%s/dns_records", Config.DomainZoneID)
	ttl := Config.RecordTTL
	if ttl == 0 {
		ttl = 1 // 1 is Automatic
	}

	var requestBody RecordData
	requestBody.Type = recordType
	requestBody.Name = Config._Name
	requestBody.Content = IP
	requestBody.TTL = ttl
	requestBody.Proxied = Config.IsProxied

	requestData, _ := json.Marshal((requestBody))
	resp := sendRequest(path, "POST", requestData)

	success, ok := resp.S("success").Data().(bool)
	if !ok {
		log.Error("[createRecord] Error decoding response")
	}

	if !success {
		errorMessage, _ := resp.S("errors").Index(0).Path("message").Data().(string)
		log.WithFields(log.Fields{"errorMessage": errorMessage}).Info("[createRecord] Failed to create record")
		return
	}
	log.WithFields(log.Fields{"recordType": IP}).Info("record created successfully")
}

func updateIP(IPversion string) {
	recordType := ""
	switch IPversion {
	case "v4":
		recordType = "A"
	case "v6":
		recordType = "AAAA"
	default:
		log.WithField("IPversion", IPversion).Error("[updateIP] Invalid IP Version")
		return
	}

	// The device's public address
	IP := getIP(IPversion)

	if IP == "" {
		// fmt.Printf("%sNo IP%s address found%s\n", color.Red, IPversion, color.Red)
		log.WithFields(log.Fields{"Version": IPversion}).Info("No IP address found")
		return
	}

	domainIP, recordID, err := getCurrentValue(recordType) // returns "" when there is no value

	if err != nil && domainIP == "" && recordID == "" {
		// create the record
		// fmt.Printf("%sIP%s address detected for the first time: %s%s\n", color.Purple, IPversion, color.Reset, IP)
		log.WithFields(log.Fields{"Version": IPversion, "IP": IP}).Info("IP address detected for the first time")
		createRecord(recordType, IP)
		runUpdateScript(IPversion, domainIP, IP)
		return
	}

	if err != nil {
		log.WithFields(log.Fields{"err": err, "recordType": recordType, "domainIP": domainIP, "recordID": recordID}).Info("[updateIP] Error getting the domain's record")
		return
	}

	if domainIP != IP {
		// fmt.Printf("%sIP%s address changed: %s%s %s->%s %s\n", color.Purple, IPversion, color.Reset, domainIP, color.Purple, color.Reset, IP)
		log.WithFields(log.Fields{"Version": IPversion, "from": domainIP, "to": IP}).Info("IP address changed")
		updateRecord(recordID, recordType, IP)
		runUpdateScript(IPversion, domainIP, IP)
		return
	}

	// fmt.Printf("%sIP%s address has not changed: %s%s\n", color.Green, IPversion, color.Reset, IP)
	log.WithFields(log.Fields{"Version": IPversion, "ip": IP}).Info("IP address has not changed")
}

func main() {
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	if *configPath == "" {
		log.Fatal("[main] config flag missing. use --config path/to/config.yaml")
		return
	}

	Config.get(*configPath)

	setupLogOutput()

	if Config.LogLevel != "" {
		level, err := log.ParseLevel(Config.LogLevel)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "DebugLevel": Config.LogLevel}).Error("[main] LogLevel has an invalid value")
		} else {
			log.Debug("Log Level set to ", Config.LogLevel)
			log.SetLevel(level)
		}
	}

	if Config.APIKey == "" {
		log.Fatal("No APIkey found in config.yaml")
	}

	if Config.Domain == "" {
		log.Fatal("No Domain found in config.yaml")
	}

	if Config.SubDomainToUpdate == "" {
		log.Warnf("No Subdomain Specified. Using root domain (%s)\n", Config.Domain)
	}

	if Config.DisableIPv4 && Config.DisableIPv6 {
		log.Fatal("IPv4 and IPv6 can't be disabled at the same time")
	}

	// fmt.Printf("%s[%s%s%s] Checking %s%s\n", color.Cyan, color.Reset, time.Now().Format(time.RFC3339), color.Cyan, color.Reset, Config.Name)
	// log.Printf("Checking %s", Config._Name)
	httpClient = &http.Client{}

	if !Config.DisableIPv4 {
		updateIP("v4")
	}

	if !Config.DisableIPv6 {
		updateIP("v6")
	}

	httpClient.CloseIdleConnections()
}

// Todo:
// * Create install script that compiles and creates systemd timer
