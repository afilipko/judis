package utils

import (
	"os"

	log "github.com/inconshreveable/log15"
)

// LogError log wrapper for errors
func LogError(msg string, err error) {
	if err != nil {
		log.Error(msg, err)
	}
}

// LogErrorAndExit log wrapper for errors which terminate judis
func LogErrorAndExit(msg string, err error) {
	if err != nil {
		log.Crit(msg, err)
		os.Exit(1)
	}
}
