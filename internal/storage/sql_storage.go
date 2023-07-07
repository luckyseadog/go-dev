package storage

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"github.com/luckyseadog/go-dev/internal/metrics"
)

type SQLStorage struct {
	DB *sql.DB
	mu sync.RWMutex
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
	ss.mu.RLock()
	defer ss.mu.RUnlock()
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
	ss.mu.RLock()
	defer ss.mu.RUnlock()
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
	ss.mu.RLock()
	defer ss.mu.RUnlock()
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
	ss.mu.RLock()
	defer ss.mu.RUnlock()
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

func NewSQLStorage(db *sql.DB) *SQLStorage {
	return &SQLStorage{DB: db, mu: sync.RWMutex{}}
}

//func (ss *SQLStorage) StoreList(metricsList []metrics.Metrics) error {
//	ss.mu.RLock()
//	defer ss.mu.RUnlock()
//
//	tx, err := ss.DB.Begin()
//	if err != nil {
//		return err
//	}
//	defer tx.Rollback()
//
//	stmtGauge, err := tx.PrepareContext(context.Background(), `
//       INSERT INTO gauge (metric, val)
//       VALUES ($1, $2)
//       ON CONFLICT (metric)
//       DO UPDATE SET val = EXCLUDED.val;
//   `)
//	if err != nil {
//		return err
//	}
//	defer stmtGauge.Close()
//
//	stmtCounter, err := tx.PrepareContext(context.Background(), `
//       INSERT INTO counter (metric, val)
//       VALUES ($1, $2)
//       ON CONFLICT (metric)
//       DO UPDATE SET val = counter.val + EXCLUDED.val;
//   `)
//	if err != nil {
//		return err
//	}
//	defer stmtCounter.Close()
//
//	for _, metric := range metricsList {
//		if metric.Value != nil {
//			if _, err = stmtGauge.ExecContext(context.Background(), stmtGauge, metric.ID, metric.Value); err != nil {
//				return err
//			}
//		} else if metric.Delta != nil {
//			if _, err = stmtCounter.ExecContext(context.Background(), stmtGauge, metric.ID, metric.Delta); err != nil {
//				return err
//			}
//		} else {
//			return errNoData
//		}
//	}
//	return tx.Commit()
//}

//func (s *MyStorage) LoadFromDB(db *sql.DB) error {
//	rowsGauge, err := db.QueryContext(context.Background(), `SELECT metric, val FROM gauge`)
//	if err != nil {
//		return err
//	}
//	defer rowsGauge.Close()
//
//	for rowsGauge.Next() {
//		var metric metrics.Metric
//		var val metrics.Gauge
//
//		err = rowsGauge.Scan(&metric, &val)
//		if err != nil {
//			return err
//		}
//		err = s.Store(metric, val)
//		if err != nil {
//			return err
//		}
//	}
//
//	if rowsGauge.Err() != nil {
//		return rowsGauge.Err()
//	}
//
//	rowsCounter, err := db.QueryContext(context.Background(), `SELECT metric, val FROM counter`)
//	if err != nil {
//		return err
//	}
//	defer rowsCounter.Close()
//
//	for rowsCounter.Next() {
//		var metric metrics.Metric
//		var val metrics.Counter
//
//		err = rowsCounter.Scan(&metric, &val)
//		if err != nil {
//			return err
//		}
//		err = s.Store(metric, val)
//		if err != nil {
//			return err
//		}
//	}
//
//	if rowsCounter.Err() != nil {
//		return rowsCounter.Err()
//	}
//
//	return nil
//}
//
//func (s *MyStorage) SaveToDB(db *sql.DB) error {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//
//	queryGauge := `
//       INSERT INTO gauge (metric, val)
//       VALUES ($1, $2)
//       ON CONFLICT (metric)
//       DO UPDATE SET val = EXCLUDED.val;
//   `
//
//	for metric, val := range s.DataGauge {
//		_, err := db.ExecContext(context.Background(), queryGauge, metric, val)
//		if err != nil {
//			return err
//		}
//	}
//
//	queryCounter := `
//       INSERT INTO counter (metric, val)
//       VALUES ($1, $2)
//       ON CONFLICT (metric)
//       DO UPDATE SET val = EXCLUDED.val;
//   `
//
//	for metric, val := range s.DataCounter {
//		_, err := db.ExecContext(context.Background(), queryCounter, metric, val)
//		if err != nil {
//			return err
//		}
//	}
//	return nil
//}
