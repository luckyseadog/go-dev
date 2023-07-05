package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/luckyseadog/go-dev/internal/metrics"
	"sync"
)

type SqlStorage struct {
	DB *sql.DB
	mu sync.RWMutex
}

func (ss *SqlStorage) CreateTables() error {
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

func (ss *SqlStorage) Store(metric metrics.Metric, metricValue any) error {
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
		_, err := ss.DB.ExecContext(context.Background(), queryGauge, metric, metricValue)
		if err != nil {
			return err
		}
		return nil
	case float64:
		_, err := ss.DB.ExecContext(context.Background(), queryGauge, metric, metricValue)
		if err != nil {
			return err
		}
		return nil
	case metrics.Counter:
		_, err := ss.DB.ExecContext(context.Background(), queryCounter, metric, metricValue)
		if err != nil {
			return err
		}
		return nil
	case int64:
		_, err := ss.DB.ExecContext(context.Background(), queryCounter, metric, metricValue)
		if err != nil {
			return err
		}
		return nil
	default:
		return errNotExpectedType
	}
}

func (ss *SqlStorage) Load(metricType string, metric metrics.Metric) (any, error) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	if metricType == "gauge" {
		var valueGauge metrics.Gauge
		row := ss.DB.QueryRowContext(context.Background(), `SELECT val FROM gauge WHERE gauge.metric = $1`, metric)
		err := row.Scan(&valueGauge)
		if err != nil {
			return nil, errors.New("no such metric")
		}
		return valueGauge, nil
	} else if metricType == "counter" {
		var valueCounter metrics.Counter
		row := ss.DB.QueryRowContext(context.Background(), `SELECT val FROM counter WHERE counter.metric = $1`, metric)
		err := row.Scan(&valueCounter)
		if err != nil {
			return nil, errors.New("no such metric")
		}
		return valueCounter, nil
	} else {
		return nil, errors.New("no such metric")
	}
}

func (ss *SqlStorage) LoadDataGauge() (map[metrics.Metric]metrics.Gauge, error) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	rowsGauge, err := ss.DB.QueryContext(context.Background(), `SELECT metric, val FROM gauge`)
	copyDataGauge := make(map[metrics.Metric]metrics.Gauge)
	if err != nil {
		return nil, err
	}
	defer rowsGauge.Close()

	for rowsGauge.Next() {
		var metric metrics.Metric
		var val metrics.Gauge

		err = rowsGauge.Scan(&metric, &val)
		if err != nil {
			return nil, err
		}
		copyDataGauge[metric] = val
	}

	if rowsGauge.Err() != nil {
		return nil, rowsGauge.Err()
	}
	return copyDataGauge, nil
}

func (ss *SqlStorage) LoadDataCounter() (map[metrics.Metric]metrics.Counter, error) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	rowsCounter, err := ss.DB.QueryContext(context.Background(), `SELECT metric, val FROM counter`)
	copyDataCounter := make(map[metrics.Metric]metrics.Counter)
	if err != nil {
		return nil, err
	}
	defer rowsCounter.Close()

	for rowsCounter.Next() {
		var metric metrics.Metric
		var val metrics.Counter

		err = rowsCounter.Scan(&metric, &val)
		if err != nil {
			return nil, err
		}
		copyDataCounter[metric] = val
	}

	if rowsCounter.Err() != nil {
		return nil, rowsCounter.Err()
	}
	return copyDataCounter, nil
}

func NewSqlStorage(db *sql.DB) *SqlStorage {
	return &SqlStorage{DB: db, mu: sync.RWMutex{}}
}

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
