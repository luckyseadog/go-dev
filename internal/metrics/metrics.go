package metrics

import (
	"runtime"
)

type Gauge float64

type Counter int64

type Metric string

var MapMetricTypes = map[string]string{
	"Alloc":         "Gauge",
	"BuckHashSys":   "Gauge",
	"Frees":         "Gauge",
	"GCCPUFraction": "Gauge",
	"GCSys":         "Gauge",
	"HeapAlloc":     "Gauge",
	"HeapIdle":      "Gauge",
	"HeapInuse":     "Gauge",
	"HeapObjects":   "Gauge",
	"HeapReleased":  "Gauge",
	"HeapSys":       "Gauge",
	"LastGC":        "Gauge",
	"Lookups":       "Gauge",
	"MCacheInuse":   "Gauge",
	"MCacheSys":     "Gauge",
	"MSpanInuse":    "Gauge",
	"MSpanSys":      "Gauge",
	"Mallocs":       "Gauge",
	"NextGC":        "Gauge",
	"NumForcedGC":   "Gauge",
	"NumGC":         "Gauge",
	"OtherSys":      "Gauge",
	"PauseTotalNs":  "Gauge",
	"StackInuse":    "Gauge",
	"StackSys":      "Gauge",
	"Sys":           "Gauge",
	"TotalAlloc":    "Gauge",
	"PollCount":     "Counter",
}

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
}
