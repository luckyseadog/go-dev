package server

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type EnvVariables struct {
	Address       string
	StoreInterval time.Duration
	StoreFile     string
	Restore       bool
}

func SetUp() *EnvVariables {
	address := os.Getenv("ADDRESS")
	if address == "" {
		address = "127.0.0.1:8080"
	}

	var storeInterval time.Duration
	storeIntervalStr := os.Getenv("STORE_INTERVAL")
	if storeIntervalStr == "" {
		storeInterval = time.Second * 10
	} else if storeIntervalStr[len(storeIntervalStr)-1] == 's' {
		var err error
		storeInterval, err = time.ParseDuration(storeIntervalStr)
		if err != nil {
			log.Fatal("Invalid storeInterval")
		}
	} else {
		var err error
		numSec, err := strconv.Atoi(storeIntervalStr)
		if err != nil {
			log.Fatal("Invalid storeInterval")
		}
		storeInterval = time.Second * time.Duration(numSec)
	}

	storeFile := os.Getenv("STORE_FILE")
	if storeFile == "" {
		storeFile = "/tmp/devops-metrics-db.json"
	}

	var restore bool
	restoreStr := os.Getenv("RESTORE")
	if restoreStr == "" {
		restore = true
	} else {
		if strings.ToLower(restoreStr) == "true" {
			restore = true
		} else if strings.ToLower(restoreStr) == "false" {
			restore = false
		} else {
			log.Fatal("Invalid restore")
		}
	}

	return &EnvVariables{Address: address, StoreInterval: storeInterval, StoreFile: storeFile, Restore: restore}

}
