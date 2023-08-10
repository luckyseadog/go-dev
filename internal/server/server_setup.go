package server

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type EnvVariables struct {
	Address        string
	StoreInterval  time.Duration
	StoreFile      string
	Restore        bool
	Dir            string
	SecretKey      []byte
	DataSourceName string
	Logging          bool
}

func SetUp() (*EnvVariables, error) {
	var addressFlag string
	var storeIntervalStrFlag string
	var storeFileFlag string
	var restoreStrFlag string
	var secretKeyFlag string
	var dataSourceNameFlag string
	var logging bool

	flag.StringVar(&addressFlag, "a", "127.0.0.1:8080", "address of server")
	flag.StringVar(&storeIntervalStrFlag, "i", "300", "time to make new write in disk")
	flag.StringVar(&storeFileFlag, "f", "/tmp/devops-metrics-db.json", "file in which we are saving metrics")
	flag.StringVar(&restoreStrFlag, "r", "true", "if it is needed to load metrics from the past")
	flag.StringVar(&secretKeyFlag, "k", "", "secret key for digital signature")
	flag.StringVar(&dataSourceNameFlag, "d", "", "for accessing the underlying datastore")
	flag.BoolVar(&logging, "log", false, "whether to save log to file")
	flag.Parse()

	address := os.Getenv("ADDRESS")
	if address == "" {
		if addressFlag == "" {
			return nil, errors.New("address can not be empty")
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
		return nil, errors.New("invalid storeInterval")
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
			return nil, errors.New("invalid restore")
		}
	}

	secretKeyStr := os.Getenv("KEY")
	if secretKeyStr == "" {
		secretKeyStr = secretKeyFlag
	}

	dataSourceNameStr := os.Getenv("DATABASE_DSN")
	if dataSourceNameStr == "" {
		dataSourceNameStr = dataSourceNameFlag
	}

	envVariables := &EnvVariables{Address: address,
		StoreInterval:  storeInterval,
		StoreFile:      storeFile,
		Restore:        restore,
		Dir:            filepath.Dir(storeFile),
		SecretKey:      []byte(secretKeyStr),
		DataSourceName: dataSourceNameStr,
		Logging:        logging,
	}

	if _, err := os.Stat(envVariables.Dir); os.IsNotExist(err) {
		err := os.Mkdir(envVariables.Dir, 0777)
		if err != nil {
			return nil, err
		}
	}

	return envVariables, nil

}
