package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/TwiN/go-color"
	"gopkg.in/yaml.v2"
)

const (
	baseURL = "https://api.cloudflare.com/client/v4/"
)

var Config conf
var httpClient *http.Client

type conf struct {
	Domain            string `yaml:"Domain" binding:"required"`
	DomainZoneID      string `yaml:"DomainZoneID"`
	SubDomainToUpdate string `yaml:"SubDomainToUpdate"`
	APIKey            string `yaml:"APIKey" binding:"required"`
	Email             string `yaml:"Email" binding:"required"`
	RecordTTL         int    `yaml:"RecordTTL" binding:"required"`
	IsProxied         bool   `yaml:"IsProxied" binding:"required"`
	DisableIPv4       bool   `yaml:"DisableIPv4"`
	DisableIPv6       bool   `yaml:"DisableIPv6"`
	Verbose           bool   `yaml:"Verbose"`
}

type RecordData struct {
	Type    string `json:"type" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Content string `json:"content" binding:"required"`
	TTL     int    `json:"ttl" binding:"required"`
	Proxied bool   `json:"proxied" binding:"required"`
}

func (c *conf) get(configPath string) *conf {
	yamlFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}

func sendRequest(path string, method string, requestBody []byte) *gabs.Container {
	url := fmt.Sprintf("%s%s", baseURL, path)
	if Config.Verbose {
		fmt.Printf("%s%s %s%s\n", color.Yellow, method, url, color.Reset)
	}

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

	req.Header.Set("X-Auth-Key", Config.APIKey)
	req.Header.Set("X-Auth-Email", Config.Email)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)

	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	if Config.Verbose {
		fmt.Printf("%s%s%s", color.Ize(color.Blue, "----- Response Starts -----\n"), string(body), color.Ize(color.Blue, "\n----- Response Ends -----\n"))
	}

	jsonParsed, err := gabs.ParseJSON(body)

	if err != nil {
		log.Fatal(("Failed to parse JSON"))
	}

	return jsonParsed
}

func getIP(ipVersion string) string {
	url := fmt.Sprintf("https://ip%s.icanhazip.com", ipVersion)

	resp, err := http.Get(url)

	if err != nil {
		// log.Fatalf("An Error Occured %v", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
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
		fmt.Println("Error decoding zoneID")
	}
	return zoneID
}

// Get the domain's current value for the specified record type (A, AAAA, TXT, etc.)
func getCurrentValue(recordType string) (string, string, error) {
	// https://api.cloudflare.com/#dns-records-for-a-zone-list-dns-records
	// name is the FQDN. 'subdomain.domain.tld' or 'domain.tld'
	zoneID := Config.DomainZoneID

	if zoneID == "" {
		fmt.Println("ZoneID not in config file, fetching from CF.")
		zoneID = getZoneID()
		Config.DomainZoneID = zoneID // save for later use but don't save to file
	}

	name := ""
	if Config.SubDomainToUpdate == "" {
		name = Config.Domain
	} else {
		name = fmt.Sprintf("%s.%s", Config.SubDomainToUpdate, Config.Domain)
	}

	path := fmt.Sprintf("zones/%s/dns_records?type=%s&name=%s", zoneID, recordType, name)
	resp := sendRequest(path, "GET", nil)

	success, ok := resp.Path("success").Data().(bool)

	if !ok {
		return "", "", errors.New("Failed to decode JSON")
	}

	if !success {
		error := resp.S("errors").Index(0)
		errorCode := error.Path("code").Data().(string)
		if errorCode == "7003" {
			return "", "", errors.New("Domain/subdomain does not exist")
		}
		return "", "", errors.New(fmt.Sprintf("errorCode: %s: %s", errorCode, error.Path("message").Data().(string)))
	}

	result := resp.S("result")
	resultLen, err := result.ArrayCount()

	if err != nil {
		return "", "", err
	}

	// the subdomain exists but there is no record for this type. There is an A record but no AAAA record or vice versa.
	if resultLen == 0 {
		return "", "", errors.New(fmt.Sprintf("No record of type %s for %s", recordType, name))
	}

	content, ok := result.Index(0).Path("content").Data().(string) // The record's value
	if !ok {
		return "", "", errors.New("Failed to decode the record's value from JSON")
	}

	if content == "" {
		return content, "", errors.New(fmt.Sprintf("No Content for %s's %s record", name, recordType))
	}

	RecordID, ok := result.Index(0).Path("id").Data().(string) // The id of the actual A or AAAA record, needed to update it.
	if !ok {
		return "", "", errors.New("Failed to decode the record's ID from JSON")
	}

	if RecordID == "" {
		return content, "", errors.New(fmt.Sprintf("No recordID for %s's %s record", name, recordType))
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
	name := ""
	if Config.SubDomainToUpdate == "" {
		name = Config.Domain
	} else {
		name = fmt.Sprintf("%s.%s", Config.SubDomainToUpdate, Config.Domain)
	}

	var requestBody RecordData
	requestBody.Type = recordType
	requestBody.Name = name
	requestBody.Content = IP
	requestBody.TTL = ttl
	requestBody.Proxied = Config.IsProxied

	requestData, _ := json.Marshal((requestBody))
	resp := sendRequest(path, "PUT", requestData)

	success, ok := resp.S("success").Data().(bool)
	if !ok {
		fmt.Println("Error decoding zoneID")
	}

	if !success {
		errorMessage, _ := resp.S("errors").Index(0).Path("message").Data().(string)
		fmt.Printf("Failed to update the record: %s", errorMessage)
		return
	}
	fmt.Printf("%s%s record changed successfully%s\n", color.Green, recordType, color.Reset)
}

func updateIP(IPversion string) {
	recordType := ""
	if IPversion == "v4" {
		recordType = "A"
	} else if IPversion == "v6" {
		recordType = "AAAA"
	} else {
		fmt.Println("Invalid IP Version.", IPversion)
		return
	}

	// The device's public address
	IP := getIP(IPversion)

	if IP == "" {
		fmt.Printf("%sNo IP%s address found%s\n", color.Red, IPversion, color.Red)
		return
	}

	domainIP, recordID, err := getCurrentValue(recordType) // returns "" when there is no value

	if err != nil {
		fmt.Printf("%sError getting the domain's %s%s%s record: %s%s\n", color.Red, color.Reset, recordType, color.Red, color.Reset, err)
		return
	}

	if domainIP != IP {
		fmt.Printf("%sIP%s address changed: %s%s %s->%s %s\n", color.Purple, IPversion, color.Reset, domainIP, color.Purple, color.Reset, IP)
		updateRecord(recordID, recordType, IP)
	} else {
		fmt.Printf("%sIP%s address has not changed: %s%s\n", color.Green, IPversion, color.Reset, IP)
	}
}

func main() {
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	Config.get(*configPath)
	if Config.APIKey == "" {
		log.Fatal("No APIkey found in config.yaml")
	}

	if Config.Email == "" {
		log.Fatal("No Email found in config.yaml")
	}

	if Config.Domain == "" {
		log.Fatal("No Domain found in config.yaml")
	}

	if Config.SubDomainToUpdate == "" {
		log.Printf("WARNING: No Subdomain Specified. Using root domain (%s)\n", Config.Domain)
	}

	if Config.DisableIPv4 && Config.DisableIPv6 {
		log.Fatal("IPv4 and IPv6 can't be disabled at the same time")
	}

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
// * Test it
// * Add/create domain/record if it doesn't exist
// * Create install script that compiles and creates systemd timer
// * Upload to github
