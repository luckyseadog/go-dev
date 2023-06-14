package storage

import (
	"reflect"
	"sync"
	"testing"

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
			}
			got, err := s.Load(tt.args.metric)
			require.NoError(t, err)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Load() got = %v, want %v", got, tt.want)
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
				mu:          sync.Mutex{},
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
