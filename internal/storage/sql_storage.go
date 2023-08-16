package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/luckyseadog/go-dev/internal/metrics"
)

type SQLStorage struct {
	DB *sql.DB
}

func NewSQLStorage(db *sql.DB) *SQLStorage {
	return &SQLStorage{DB: db}
}

func (ss *SQLStorage) CreateTables() error {
	_, err := ss.DB.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS gauge (
				  metric VARCHAR(100) UNIQUE,
				  val DOUBLE PRECISION
				)`)
	if err != nil {
		return err
	}

	_, err = ss.DB.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS counter (
				  metric VARCHAR(100) UNIQUE,
				  val BIGINT
				)`)
	if err != nil {
		return err
	}

	return nil
}

func (ss *SQLStorage) StoreContext(ctx context.Context, metric metrics.Metric, metricValue any) error {
	queryGauge := `
       INSERT INTO gauge (metric, val)
       VALUES ($1, $2)
       ON CONFLICT (metric)
       DO UPDATE SET val = EXCLUDED.val;
   `
	queryCounter := `
       INSERT INTO counter (metric, val)
       VALUES ($1, $2)
       ON CONFLICT (metric)
       DO UPDATE SET val = counter.val + EXCLUDED.val;
   `

	switch metricValue := metricValue.(type) {
	case metrics.Gauge:
		_, err := ss.DB.ExecContext(ctx, queryGauge, metric, metricValue)
		if err != nil {
			return err
		}
		return nil
	case float64:
		_, err := ss.DB.ExecContext(ctx, queryGauge, metric, metricValue)
		if err != nil {
			return err
		}
		return nil
	case metrics.Counter:
		_, err := ss.DB.ExecContext(ctx, queryCounter, metric, metricValue)
		if err != nil {
			return err
		}
		return nil
	case int64:
		_, err := ss.DB.ExecContext(ctx, queryCounter, metric, metricValue)
		if err != nil {
			return err
		}
		return nil
	default:
		return errNotExpectedType
	}
}

func (ss *SQLStorage) LoadContext(ctx context.Context, metricType string, metric metrics.Metric) Result {
	if metricType == "gauge" {
		var valueGauge metrics.Gauge
		row := ss.DB.QueryRowContext(ctx, `SELECT val FROM gauge WHERE gauge.metric = $1`, metric)
		err := row.Scan(&valueGauge)
		if err != nil {
			return Result{Value: nil, Err: errors.New("no such metric")}
		}
		return Result{Value: valueGauge, Err: nil}
	} else if metricType == "counter" {
		var valueCounter metrics.Counter
		row := ss.DB.QueryRowContext(ctx, `SELECT val FROM counter WHERE counter.metric = $1`, metric)
		err := row.Scan(&valueCounter)
		if err != nil {
			return Result{Value: nil, Err: errors.New("no such metric")}
		}
		return Result{Value: valueCounter, Err: nil}
	} else {
		return Result{Value: nil, Err: errors.New("no such metric")}
	}
}

func (ss *SQLStorage) LoadDataGaugeContext(ctx context.Context) Result {
	rowsGauge, err := ss.DB.QueryContext(ctx, `SELECT metric, val FROM gauge`)
	copyDataGauge := make(map[metrics.Metric]metrics.Gauge)
	if err != nil {
		return Result{Value: nil, Err: err}
	}
	defer rowsGauge.Close()

	for rowsGauge.Next() {
		var metric metrics.Metric
		var val metrics.Gauge

		err = rowsGauge.Scan(&metric, &val)
		if err != nil {
			return Result{Value: nil, Err: err}
		}
		copyDataGauge[metric] = val
	}

	if rowsGauge.Err() != nil {
		return Result{Value: nil, Err: rowsGauge.Err()}
	}
	return Result{Value: copyDataGauge, Err: nil}
}

func (ss *SQLStorage) LoadDataCounterContext(ctx context.Context) Result {
	rowsCounter, err := ss.DB.QueryContext(ctx, `SELECT metric, val FROM counter`)
	copyDataCounter := make(map[metrics.Metric]metrics.Counter)
	if err != nil {
		return Result{Value: nil, Err: err}
	}
	defer rowsCounter.Close()

	for rowsCounter.Next() {
		var metric metrics.Metric
		var val metrics.Counter

		err = rowsCounter.Scan(&metric, &val)
		if err != nil {
			return Result{Value: nil, Err: err}
		}
		copyDataCounter[metric] = val
	}

	if rowsCounter.Err() != nil {
		return Result{Value: nil, Err: rowsCounter.Err()}
	}
	return Result{Value: copyDataCounter, Err: nil}
}
