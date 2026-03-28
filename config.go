package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/goccy/go-yaml"
)

type Config struct {
	// Complete FQDN to update. Set by the program.
	_Name string
	//  The domain name to update
	Domain string `yaml:"Domain" binding:"required"`
	// The Cloudflare Zone ID for the Domain. If left empty, it will be fetched from Cloudflare. Setting it removes the need for an extra API call.
	DomainZoneID string `yaml:"DomainZoneID"`
	// The subdomain of the Domain to update. If left empty, the Domain itself is used.
	SubDomainToUpdate string `yaml:"SubDomainToUpdate"`
	// The Cloudflare Account Token with DNS Read and Edit permissions.
	// Create Token: https://developers.cloudflare.com/fundamentals/api/get-started/create-token/
	APIKey string `yaml:"APIKey" binding:"required"`
	// The TTL assigned to the domain in seconds. 1 sets it to cloudflare's automatic option.
	RecordTTL int `yaml:"RecordTTL"`
	// Use Cloudflare to proxy your traffic. Equivalent to enabling the cloud in Cloudflare.
	IsProxied bool `yaml:"IsProxied"`
	// Disable checking and updating IPv4 and A Records
	DisableIPv4 bool `yaml:"DisableIPv4"`
	// Disable checking and updating IPv6 and AAAA Records
	DisableIPv6 bool `yaml:"DisableIPv6"`
	//  The path to a script or binary that gets executed when the IP address changes.
	// The arguments are: the IP version ("v4" or "v6"), the old IP, the new IP, and the updated FQDN in that order.
	ScriptOnChange string `yaml:"ScriptOnChange"`
	// The path to a script or binary that gets executed when there is an error updating the IP Address. It only gets called if updating or creating a record fails.
	// It does not get called if the program is not able to get the current IP.
	// The arguments are: the error, the IP version ("v4" or "v6"), the old IP, the new IP, and the updated FQDN in that order.
	ScriptOnError string `yaml:"ScriptOnError"`
	// The path to a file to save logs to. To log to stdout, set it to'stdout'. Log library defaults to stderr.
	LogFile string `yaml:"LogFile"`
	// The level of details to log. The options from less detail to very detailed are: panic, fatal, error, warning, info, debug, and trace
	LogLevel string `yaml:"LogLevel"`
}

func (c *Config) get(configPath string) *Config {
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	c._Name = ""
	if c.SubDomainToUpdate == "" {
		c._Name = c.Domain
	} else {
		c._Name = fmt.Sprintf("%s.%s", c.SubDomainToUpdate, c.Domain)
	}

	log.WithFields(log.Fields{"Domain": c.Domain, "SubDomainToUpdate": c.SubDomainToUpdate, "APIKey": c.APIKey, "RecordTTL": c.RecordTTL, "IsProxied": c.IsProxied, "DisableIPv4": c.DisableIPv4, "DisableIPv6": c.DisableIPv6, "ScriptOnChange": c.ScriptOnChange, "LogFile": c.LogFile, "LogLevel": c.LogLevel}).Trace("Config options")

	return c
}

func setupLogOutput() {
	if conf.LogFile == "" {
		log.Info("[setupLogOutput] No LogFile specified. Logging to stderr")
		// https://pkg.go.dev/github.com/sirupsen/logrus#New
		return
	}

	if conf.LogFile == "stdout" {
		log.SetOutput(os.Stdout)
		return
	}

	// If the file doesn't exist, create it, otherwise append to the file
	file, err := os.OpenFile(conf.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "logFilePath": conf.LogFile}).Error("[setupLogOutput] Failed to open log file. Using stderr instead")
		return
	}

	log.SetOutput(file)
}
