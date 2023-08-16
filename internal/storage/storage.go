package storage

import (
	"context"
	"errors"
	"time"

	"github.com/luckyseadog/go-dev/internal/metrics"
)

var (
	ErrNotMyStorage    = errors.New("database is not of the type MyStorage")
	ErrNotSQLStorage   = errors.New("database is not of the type SQLStorage")
	errNotExpectedType = errors.New("not expected type")
	errNoSuchMetric    = errors.New("no such metric")
)

type AutoSavingParams struct {
	storageChan   chan struct{}
	storeInterval time.Duration
}

type Result struct {
	Value any
	Err   error
}

type Storage interface {
	StoreContext(ctx context.Context, metric metrics.Metric, metricValue any) error
	LoadContext(ctx context.Context, metricType string, metric metrics.Metric) Result
	LoadDataGaugeContext(ctx context.Context) Result
	LoadDataCounterContext(ctx context.Context) Result
}
