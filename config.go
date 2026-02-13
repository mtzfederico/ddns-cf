package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/goccy/go-yaml"
)

type conf struct {
	// Complete FQDN to update. Set by the program.
	_Name             string
	Domain            string `yaml:"Domain" binding:"required"`
	DomainZoneID      string `yaml:"DomainZoneID"`
	SubDomainToUpdate string `yaml:"SubDomainToUpdate"`
	APIKey            string `yaml:"APIKey" binding:"required"`
	RecordTTL         int    `yaml:"RecordTTL"`
	IsProxied         bool   `yaml:"IsProxied"`
	DisableIPv4       bool   `yaml:"DisableIPv4"`
	DisableIPv6       bool   `yaml:"DisableIPv6"`
	ScriptOnChange    string `yaml:"ScriptOnChange"`
	LogFile           string `yaml:"LogFile"`
	// The level of details to log. The options from less detail to very detailed are: panic, fatal, error, warning, info, debug, and trace
	LogLevel string `yaml:"LogLevel"`
}

func (c *conf) get(configPath string) *conf {
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

	log.WithFields(log.Fields{"Domain": Config.Domain, "SubDomainToUpdate": Config.SubDomainToUpdate, "APIKey": Config.APIKey, "RecordTTL": Config.RecordTTL, "IsProxied": Config.IsProxied, "DisableIPv4": Config.DisableIPv4, "DisableIPv6": Config.DisableIPv6, "ScriptOnChange": Config.ScriptOnChange, "LogFile": Config.LogFile, "LogLevel": Config.LogLevel}).Trace("Config options")

	return c
}

func setupLogOutput() {
	if Config.LogFile == "" {
		log.Info("[setupLogOutput] No LogFile specified. Logging to stderr")
		// https://pkg.go.dev/github.com/sirupsen/logrus#New
		return
	}

	if Config.LogFile == "stdout" {
		log.SetOutput(os.Stdout)
		return
	}

	// If the file doesn't exist, create it, otherwise append to the file
	file, err := os.OpenFile(Config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "logFilePath": Config.LogFile}).Error("[setupLogOutput] Failed to open log file. Using stderr instead")
		return
	}

	log.SetOutput(file)
}
