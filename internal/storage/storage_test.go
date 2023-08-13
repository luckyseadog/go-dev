package storage

import (
	"context"
	"path"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/luckyseadog/go-dev/internal/metrics"
	"github.com/stretchr/testify/require"
)

func TestStorage_Load(t *testing.T) {
	type fields struct {
		dataGauge   map[metrics.Metric]metrics.Gauge
		dataCounter map[metrics.Metric]metrics.Counter
	}
	type args struct {
		metric metrics.Metric
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    any
		wantErr bool
	}{
		{
			name: "LOAD",
			fields: fields{
				dataGauge: map[metrics.Metric]metrics.Gauge{
					"StackSys": 1.0,
				},
				dataCounter: map[metrics.Metric]metrics.Counter{
					"RandomValue": 1,
				},
			},
			args: args{"StackSys"},
			want: metrics.Gauge(1.0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MyStorage{
				DataGauge:   tt.fields.dataGauge,
				DataCounter: tt.fields.dataCounter,
				autoSavingParams: AutoSavingParams{
					storageChan:   nil,
					storeInterval: time.Second,
				},
			}
			res := s.Load("gauge", tt.args.metric)
			require.NoError(t, res.Err)
			if !reflect.DeepEqual(res.Value, tt.want) {
				t.Errorf("Load() got = %v, want %v", res.Value, tt.want)
			}
		})
	}
}

func TestStorage_Store(t *testing.T) {
	type fields struct {
		dataGauge   map[metrics.Metric]metrics.Gauge
		dataCounter map[metrics.Metric]metrics.Counter
	}
	type args struct {
		metric      metrics.Metric
		metricValue any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "STORE",
			fields: fields{
				dataGauge: map[metrics.Metric]metrics.Gauge{
					"StackSys": 1.0,
				},
				dataCounter: map[metrics.Metric]metrics.Counter{
					"RandomValue": 1,
				},
			},
			args: args{metric: "RandomValue", metricValue: metrics.Counter(6)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MyStorage{
				DataGauge:   tt.fields.dataGauge,
				DataCounter: tt.fields.dataCounter,
				mu:          sync.RWMutex{},
				autoSavingParams: AutoSavingParams{
					storageChan:   nil,
					storeInterval: time.Second,
				},
			}
			_ = s.Store(tt.args.metric, tt.args.metricValue)
			require.Equal(t, s.DataCounter["RandomValue"], metrics.Counter(7))
			_ = s.Store(tt.args.metric, tt.args.metricValue)
			require.Equal(t, s.DataCounter["RandomValue"], metrics.Counter(13))
			_ = s.Store(tt.args.metric, tt.args.metricValue)
			require.Equal(t, s.DataCounter["RandomValue"], metrics.Counter(19))
		})
	}
}

func TestStorage_File(t *testing.T) {
	storage := NewStorage(nil, time.Millisecond)
	err := storage.StoreContext(context.Background(), "Metric1", int64(1))
	require.NoError(t, err)
	err = storage.StoreContext(context.Background(), "Metric2", 2.0)
	require.NoError(t, err)
	err = storage.StoreContext(context.Background(), "Metric3", metrics.Counter(3))
	require.NoError(t, err)
	err = storage.StoreContext(context.Background(), "Metric4", metrics.Gauge(4.0))
	require.NoError(t, err)
	err = storage.StoreContext(context.Background(), "Metric5", "unknown")
	require.Error(t, err)

	tmpDir := t.TempDir()
	err = storage.SaveToFile(path.Join(tmpDir, "metrics.json"))
	require.NoError(t, err)

	storage = NewStorage(nil, time.Millisecond)
	err = storage.LoadFromFile(path.Join(tmpDir, "metrics.json"))
	require.NoError(t, err)

	res := storage.LoadContext(context.Background(), "counter", "Metric1")
	require.NoError(t, res.Err)
	resCounter, ok := res.Value.(metrics.Counter)
	require.True(t, ok)
	require.Equal(t, resCounter, metrics.Counter(1))

	res = storage.LoadContext(context.Background(), "gauge", "Metric2")
	require.NoError(t, res.Err)
	resGauge, ok := res.Value.(metrics.Gauge)
	require.True(t, ok)
	require.Equal(t, resGauge, metrics.Gauge(2.0))

	res = storage.LoadContext(context.Background(), "counter", "Metric3")
	require.NoError(t, res.Err)
	resCounter, ok = res.Value.(metrics.Counter)
	require.True(t, ok)
	require.Equal(t, resCounter, metrics.Counter(3))

	res = storage.LoadContext(context.Background(), "gauge", "Metric4")
	require.NoError(t, res.Err)
	resGauge, ok = res.Value.(metrics.Gauge)
	require.True(t, ok)
	require.Equal(t, resGauge, metrics.Gauge(4.0))

	res = storage.LoadContext(context.Background(), "counter", "Metric4")
	require.Error(t, res.Err)

}

func TestStorage_LoadAllData(t *testing.T) {
	dataGauge := map[metrics.Metric]metrics.Gauge{
		"metricGauge1": metrics.Gauge(1.0),
		"metricGauge2": metrics.Gauge(2.0),
	}
	dataCounter := map[metrics.Metric]metrics.Counter{
		"metricCounter1": metrics.Counter(1),
		"metricCounter2": metrics.Counter(2),
	}

	storage := MyStorage{DataGauge: dataGauge, DataCounter: dataCounter}
	res := storage.LoadDataGaugeContext(context.Background())
	require.NoError(t, res.Err)
	require.True(t, reflect.DeepEqual(res.Value, dataGauge))
	res = storage.LoadDataCounterContext(context.Background())
	require.NoError(t, res.Err)
	require.True(t, reflect.DeepEqual(res.Value, dataCounter))
}
