package server

import (
	"flag"
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
	var addressFlag string
	var storeIntervalStrFlag string
	var storeFileFlag string
	var restoreStrFlag string

	flag.StringVar(&addressFlag, "a", "127.0.0.1:8080", "address of server")
	flag.StringVar(&storeIntervalStrFlag, "i", "300", "time to make new write in disk")
	flag.StringVar(&storeFileFlag, "f", "/tmp/devops-metrics-db.json", "file in which we are saving metrics")
	flag.StringVar(&restoreStrFlag, "r", "true", "if it is needed to load metrics from the past")
	flag.Parse()

	address := os.Getenv("ADDRESS")
	if address == "" {
		if addressFlag == "" {
			log.Fatal("Address can not be empty")
		}
		address = addressFlag
	}

	log.Println(address)

	var storeInterval time.Duration
	storeIntervalStr := os.Getenv("STORE_INTERVAL")
	if storeIntervalStr == "" {
		storeIntervalStr = storeIntervalStrFlag
	}
	if storeIntervalStr == "" {
		storeInterval = 0
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
		storeFile = storeFileFlag
	}

	var restore bool
	restoreStr := os.Getenv("RESTORE")
	if restoreStr == "" {
		restoreStr = restoreStrFlag
	}
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
