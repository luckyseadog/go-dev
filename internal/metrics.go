package internal

import "runtime"

type Gauge float64

type Counter int64

func GetMetrics(memStats runtime.MemStats) map[string]Gauge {
	metricsMap := map[string]Gauge{
		"Alloc":         Gauge(memStats.Alloc),
		"BuckHashSys":   Gauge(memStats.BuckHashSys),
		"Frees":         Gauge(memStats.Frees),
		"GCCPUFraction": Gauge(memStats.GCCPUFraction),
		"GCSys":         Gauge(memStats.GCSys),
		"HeapAlloc":     Gauge(memStats.HeapAlloc),
		"HeapIdle":      Gauge(memStats.HeapIdle),
		"HeapInuse":     Gauge(memStats.HeapInuse),
		"HeapObjects":   Gauge(memStats.HeapObjects),
		"HeapReleased":  Gauge(memStats.HeapReleased),
		"HeapSys":       Gauge(memStats.HeapSys),
		"LastGC":        Gauge(memStats.LastGC),
		"Lookups":       Gauge(memStats.Lookups),
		"MCacheInuse":   Gauge(memStats.MCacheInuse),
		"MCacheSys":     Gauge(memStats.MCacheSys),
		"MSpanInuse":    Gauge(memStats.MSpanInuse),
		"MSpanSys":      Gauge(memStats.MSpanSys),
		"Mallocs":       Gauge(memStats.Mallocs),
		"NextGC":        Gauge(memStats.NextGC),
		"NumForcedGC":   Gauge(memStats.NumForcedGC),
		"NumGC":         Gauge(memStats.NumGC),
		"OtherSys":      Gauge(memStats.OtherSys),
		"PauseTotalNs":  Gauge(memStats.PauseTotalNs),
		"StackInuse":    Gauge(memStats.StackInuse),
		"StackSys":      Gauge(memStats.StackSys),
		"Sys":           Gauge(memStats.Sys),
		"TotalAlloc":    Gauge(memStats.TotalAlloc),
	}

	return metricsMap
}
