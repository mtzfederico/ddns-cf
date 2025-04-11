package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

type conf struct {
	// Complete FQDN to update. Set by the program.
	Name              string
	Domain            string `yaml:"Domain" binding:"required"`
	DomainZoneID      string `yaml:"DomainZoneID"`
	SubDomainToUpdate string `yaml:"SubDomainToUpdate"`
	APIKey            string `yaml:"APIKey" binding:"required"`
	Email             string `yaml:"Email" binding:"required"`
	RecordTTL         int    `yaml:"RecordTTL"`
	IsProxied         bool   `yaml:"IsProxied"`
	DisableIPv4       bool   `yaml:"DisableIPv4"`
	DisableIPv6       bool   `yaml:"DisableIPv6"`
	Verbose           bool   `yaml:"Verbose"`
	ScriptOnChange    string `yaml:"ScriptOnChange"`
	LogFile           string `yaml:"LogFile"`
	DebugLevel        string `yaml:"DebugLevel"`
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

	c.Name = ""
	if c.SubDomainToUpdate == "" {
		c.Name = c.Domain
	} else {
		c.Name = fmt.Sprintf("%s.%s", c.SubDomainToUpdate, c.Domain)
	}

	if c.LogFile == "" {
		c.LogFile = "/var/log/ddns-cf/ddns-cf.log"
		fmt.Println("[WARNING] Using default logging path (/var/log/ddns-cf/ddns-cf.log)")
	}

	if c.Verbose {
		fmt.Printf("Config: Domain: %s, SubDomainToUpdate. %s, APIKey: %s, Email: %s, RecordTTL: %d, IsProxied: %t, DisableIPv4: %t, DisableIPv6: %t, Verbose: %t, ScriptOnChange: %s, LogFile: %s, DebugLevel: %s\n", c.Domain, c.SubDomainToUpdate, c.APIKey, c.Email, c.RecordTTL, c.IsProxied, c.DisableIPv4, c.DisableIPv6, c.Verbose, c.ScriptOnChange, c.LogFile, c.DebugLevel)
	}

	return c
}
