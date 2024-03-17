package server

import (
	"encoding/json"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/luckyseadog/go-dev/internal/metrics"
)

type EnvVariables struct {
	Address        string
	StoreInterval  time.Duration
	StoreFile      string
	Restore        bool
	Dir            string
	SecretKey      []byte
	DataSourceName string
	Logging        bool
	CryptoKeyDir   string
	TrustedSubnet  string
	GRPC           bool
}

func SetUp() (*EnvVariables, error) {
	var addressFlag string
	var storeIntervalStrFlag string
	var storeFileFlag string
	var restoreStrFlag string
	var secretKeyFlag string
	var dataSourceNameFlag string
	var logging bool
	var cryptoKeyFlag string
	var configFlag string
	var cFlag string
	var trustedSubnetFlag string
	var gRPCFlag string

	flag.StringVar(&addressFlag, "a", "127.0.0.1:8080", "address of server")
	flag.StringVar(&storeIntervalStrFlag, "i", "300", "time to make new write in disk")
	flag.StringVar(&storeFileFlag, "f", "/tmp/devops-metrics-db.json", "file in which we are saving metrics")
	flag.StringVar(&restoreStrFlag, "r", "true", "if it is needed to load metrics from the past")
	flag.StringVar(&secretKeyFlag, "k", "", "secret key for digital signature")
	flag.StringVar(&dataSourceNameFlag, "d", "", "for accessing the underlying datastore")
	flag.BoolVar(&logging, "log", false, "whether to save log to file")
	flag.StringVar(&cryptoKeyFlag, "crypto-key", "", "whether to use asymmetric encoding")
	flag.StringVar(&configFlag, "config", "", "path to config")
	flag.StringVar(&cFlag, "c", "", "path to config")
	flag.StringVar(&trustedSubnetFlag, "t", "", "mask of subnet which is trusted")
	flag.StringVar(&gRPCFlag, "grpc", "false", "whether to use gRPC")
	flag.Parse()

	var configPath string
	configStr := os.Getenv("CONFIG")
	if configStr == "" {
		if configFlag != "" {
			configPath = configFlag
		} else {
			configPath = cFlag
		}
	} else {
		configPath = configStr
	}

	var Config metrics.ConfigServer
	if configPath != "" {
		f, err := os.ReadFile(configPath)
		if err != nil {
			return nil, errors.New("invalid config")
		}
		err = json.Unmarshal(f, &Config)
		if err != nil {
			return nil, errors.New("config has bad json")
		}
	}
	if addressFlag == "" {
		addressFlag = Config.Address
	}
	if storeIntervalStrFlag == "" {
		storeIntervalStrFlag = Config.StoreInterval
	}
	if storeFileFlag == "" {
		storeFileFlag = Config.StoreFile
	}
	if restoreStrFlag == "" {
		restoreStrFlag = Config.Restore
	}
	if secretKeyFlag == "" {
		secretKeyFlag = Config.SecretKey
	}
	if dataSourceNameFlag == "" {
		dataSourceNameFlag = Config.DataSourseName
	}
	if cryptoKeyFlag == "" {
		cryptoKeyFlag = Config.CryptoKey
	}

	if trustedSubnetFlag == "" {
		trustedSubnetFlag = Config.TrustedSubnet
	}

	if gRPCFlag == "" {
		gRPCFlag = Config.GRPC
	}

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

	cryptoKeyStr := os.Getenv("CRYPTO_KEY")
	if cryptoKeyStr == "" {
		cryptoKeyStr = cryptoKeyFlag
	}

	trustedSubnetStr := os.Getenv("TRUSTED_SUBNET")
	if trustedSubnetStr == "" {
		trustedSubnetStr = trustedSubnetFlag
	}

	var gRPC bool
	gRPCStr := os.Getenv("GRPC")
	if gRPCStr == "" {
		gRPCStr = gRPCFlag
	}
	if gRPCStr == "" {
		gRPC = false
	} else {
		if strings.ToLower(gRPCStr) == "true" {
			gRPC = true
		} else if strings.ToLower(gRPCStr) == "false" {
			gRPC = false
		} else {
			return nil, errors.New("invalid gRPC flag")
		}
	}

	envVariables := &EnvVariables{Address: address,
		StoreInterval:  storeInterval,
		StoreFile:      storeFile,
		Restore:        restore,
		Dir:            filepath.Dir(storeFile),
		SecretKey:      []byte(secretKeyStr),
		DataSourceName: dataSourceNameStr,
		Logging:        logging,
		CryptoKeyDir:   cryptoKeyStr,
		TrustedSubnet:  trustedSubnetStr,
		GRPC:           gRPC,
	}

	if _, err := os.Stat(envVariables.Dir); os.IsNotExist(err) {
		err := os.Mkdir(envVariables.Dir, 0777)
		if err != nil {
			return nil, err
		}
	}

	return envVariables, nil

}
