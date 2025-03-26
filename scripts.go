package main

import (
	"os/exec"

	log "github.com/sirupsen/logrus"
)

// Runs the script specified in the config file (if any) when an IP changes.
// The arguments are: IPversion, OldIP, NewIP, Updated FQDN
func runUpdateScript(IPversion string, OldIP string, NewIP string) {
	scriptPath := Config.ScriptOnChange
	if scriptPath == "" {
		log.Info("[runUpdateScript] No script found")
		return
	}

	out, err := exec.Command(scriptPath, IPversion, OldIP, NewIP, Config.Name).Output()
	if err != nil {
		log.WithFields(log.Fields{"IPversion": IPversion, "out": out, "err": err}).Error("[runUpdateScript] Error from script")
		return
	}
	log.WithFields(log.Fields{"IPversion": IPversion, "out": string(out)}).Info("[runUpdateScript] Script ran")
}
