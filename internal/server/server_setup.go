package server

import (
	"flag"
	"github.com/luckyseadog/go-dev/internal/storage"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type EnvVariables struct {
	Address       string
	StoreInterval time.Duration
	StoreFile     string
	Restore       bool
	Dir           string
}

func SetUp(s storage.Storage) *EnvVariables {
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

	var storeInterval time.Duration
	storeIntervalStr := os.Getenv("STORE_INTERVAL")
	if storeIntervalStr == "" {
		storeIntervalStr = storeIntervalStrFlag
	}
	if storeIntervalStr == "" {
		storeInterval = 0
	} else if duration, err := time.ParseDuration(storeIntervalStr); err == nil {
		storeInterval = duration
	} else if numSec, err := strconv.Atoi(storeIntervalStr); err == nil {
		storeInterval = time.Second * time.Duration(numSec)
	} else {
		log.Fatal("Invalid storeInterval")
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

	envVariables := &EnvVariables{Address: address,
		StoreInterval: storeInterval,
		StoreFile:     storeFile,
		Restore:       restore,
		Dir:           filepath.Dir(storeFile),
	}

	if _, err := os.Stat(envVariables.Dir); os.IsNotExist(err) {
		err := os.Mkdir(envVariables.Dir, 0777)
		if err != nil {
			log.Fatal(err)
		}
	}

	if envVariables.Restore {
		if _, err := os.Stat(envVariables.StoreFile); err == nil {
			err := s.LoadFromFile(envVariables.StoreFile)
			if err != nil {
				log.Println(err)
			}
		}
	}

	return envVariables

}
