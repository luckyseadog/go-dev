package metrics

import (
	"runtime"
)

type Gauge float64

type Counter int64

type Metric string

const (
	RandomValue   = Metric("RandomValue")
	PollCount     = Metric("PollCount")
	Alloc         = Metric("Alloc")
	BuckHashSys   = Metric("BuckHashSys")
	Frees         = Metric("Frees")
	GCCPUFraction = Metric("GCCPUFraction")
	GCSys         = Metric("GCSys")
	HeapAlloc     = Metric("HeapAlloc")
	HeapIdle      = Metric("HeapIdle")
	HeapInuse     = Metric("HeapInuse")
	HeapObjects   = Metric("HeapObjects")
	HeapReleased  = Metric("HeapReleased")
	HeapSys       = Metric("HeapSys")
	LastGC        = Metric("LastGC")
	Lookups       = Metric("Lookups")
	MCacheInuse   = Metric("MCacheInuse")
	MCacheSys     = Metric("MCacheSys")
	MSpanInuse    = Metric("MSpanInuse")
	MSpanSys      = Metric("MSpanSys")
	Mallocs       = Metric("Mallocs")
	NextGC        = Metric("NextGC")
	NumForcedGC   = Metric("NumForcedGC")
	NumGC         = Metric("NumGC")
	OtherSys      = Metric("OtherSys")
	PauseTotalNs  = Metric("PauseTotalNs")
	StackInuse    = Metric("StackInuse")
	StackSys      = Metric("StackSys")
	Sys           = Metric("Sys")
	TotalAlloc    = Metric("TotalAlloc")

	TotalMemory     = Metric("TotalMemory")
	FreeMemory      = Metric("FreeMemory")
	CPUutilization1 = Metric("CPUutilization1")
)

func GetMetrics(memStats runtime.MemStats) map[Metric]Gauge {
	metricsMap := map[Metric]Gauge{
		Alloc:         Gauge(memStats.Alloc),
		BuckHashSys:   Gauge(memStats.BuckHashSys),
		Frees:         Gauge(memStats.Frees),
		GCCPUFraction: Gauge(memStats.GCCPUFraction),
		GCSys:         Gauge(memStats.GCSys),
		HeapAlloc:     Gauge(memStats.HeapAlloc),
		HeapIdle:      Gauge(memStats.HeapIdle),
		HeapInuse:     Gauge(memStats.HeapInuse),
		HeapObjects:   Gauge(memStats.HeapObjects),
		HeapReleased:  Gauge(memStats.HeapReleased),
		HeapSys:       Gauge(memStats.HeapSys),
		LastGC:        Gauge(memStats.LastGC),
		Lookups:       Gauge(memStats.Lookups),
		MCacheInuse:   Gauge(memStats.MCacheInuse),
		MCacheSys:     Gauge(memStats.MCacheSys),
		MSpanInuse:    Gauge(memStats.MSpanInuse),
		MSpanSys:      Gauge(memStats.MSpanSys),
		Mallocs:       Gauge(memStats.Mallocs),
		NextGC:        Gauge(memStats.NextGC),
		NumForcedGC:   Gauge(memStats.NumForcedGC),
		NumGC:         Gauge(memStats.NumGC),
		OtherSys:      Gauge(memStats.OtherSys),
		PauseTotalNs:  Gauge(memStats.PauseTotalNs),
		StackInuse:    Gauge(memStats.StackInuse),
		StackSys:      Gauge(memStats.StackSys),
		Sys:           Gauge(memStats.Sys),
		TotalAlloc:    Gauge(memStats.TotalAlloc),
	}

	return metricsMap
}

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

type FileData struct {
	DataGauge   map[Metric]Gauge   `json:"data_gauge"`
	DataCounter map[Metric]Counter `json:"data_counter"`
}

type ConfigAgent struct {
	Address        string `json:"address,omitempty"`
	PollInterval   string `json:"poll_interval,omitempty"`
	ReportInterval string `json:"report_interval,omitempty"`
	SecretKey      string `json:"secret_key,omitempty"`
	RateLimit      string `json:"rate_limit,omitempty"`
	CryptoKey      string `json:"crypto_key,omitempty"`
	GRPC           string `json:"grpc,omitempty"`
}

type ConfigServer struct {
	Address        string `json:"address,omitempty"`
	StoreInterval  string `json:"store_interval,omitempty"`
	StoreFile      string `json:"store_file,omitempty"`
	Restore        string `json:"restore,omitempty"`
	SecretKey      string `json:"secret_key,omitempty"`
	DataSourseName string `json:"database_dsn,omitempty"`
	CryptoKey      string `json:"crypto_key,omitempty"`
	TrustedSubnet  string `json:"trusted_subnet,omitempty"`
	GRPC           string `json:"grpc,omitempty"`
}
