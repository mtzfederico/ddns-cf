package main

import (
	"os/exec"

	log "github.com/sirupsen/logrus"
)

// Runs the script specified in the config file (if any) when an IP changes.
// The arguments are: IPversion, OldIP, NewIP, Updated FQDN
func runUpdateScript(version IPVersion, oldIP string, newIP string) {
	scriptPath := conf.ScriptOnChange
	if scriptPath == "" {
		log.Info("[runUpdateScript] No script found")
		return
	}

	out, err := exec.Command(scriptPath, string(version), oldIP, newIP, conf._Name).Output()
	if err != nil {
		log.WithFields(log.Fields{"IPversion": version, "out": out, "err": err}).Error("[runUpdateScript] Error from script")
		return
	}
	log.WithFields(log.Fields{"IPversion": version, "out": string(out)}).Info("[runUpdateScript] Script ran")
}

// Runs the script specified in the config file (if any) when there is an error updating the IP Address.
// The arguments are: error, IPversion, OldIP, NewIP, Updated FQDN
func runErrorScript(err error, version IPVersion, oldIP string, newIP string) {
	scriptPath := conf.ScriptOnError
	if scriptPath == "" {
		log.Info("[runErrorScript] No script found")
		return
	}

	out, scriptErr := exec.Command(scriptPath, err.Error(), string(version), oldIP, newIP, conf._Name).Output()
	if scriptErr != nil {
		log.WithFields(log.Fields{"err": err, "IPversion": version, "out": out, "scriptErr": scriptErr}).Error("[runErrorScript] Error from script")
		return
	}
	log.WithFields(log.Fields{"out": string(out)}).Info("[runErrorScript] Script ran")
}
