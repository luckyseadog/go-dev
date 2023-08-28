// Package storage provides functionalities for keeping data on server
// this package have two types of storage:
// - MyStorage that keeps data in map and periodically saves it to file
// - SQLStorage that keeps data as SQL database
package storage

import (
	"context"
	"errors"
	"time"

	"github.com/luckyseadog/go-dev/internal/metrics"
)

// Various error variables used in the storage package.
var (
	ErrNotMyStorage    = errors.New("database is not of the type MyStorage")
	ErrNotSQLStorage   = errors.New("database is not of the type SQLStorage")
	errNotExpectedType = errors.New("not expected type")
	errNoSuchMetric    = errors.New("no such metric")
)

// AutoSavingParams is a structure that holds parameters related to auto-saving data in the storage.
type AutoSavingParams struct {
	storageChan   chan struct{}
	storeInterval time.Duration
}

// Result is a structure that holds data that Storage receive after loading data.
type Result struct {
	Value any
	Err   error
}

// Storage is an interface for some concrete storage implementation.
type Storage interface {
	StoreContext(ctx context.Context, metric metrics.Metric, metricValue any) error
	LoadContext(ctx context.Context, metricType string, metric metrics.Metric) Result
	LoadDataGaugeContext(ctx context.Context) Result
	LoadDataCounterContext(ctx context.Context) Result
}
