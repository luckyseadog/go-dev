package pkg

import (
	"github.com/luckyseadog/go-dev/internal"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestStorage_Load(t *testing.T) {
	type fields struct {
		dataGauge   map[internal.Metric]internal.Gauge
		dataCounter map[internal.Metric]*internal.Queue
	}
	type args struct {
		metric internal.Metric
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
				map[internal.Metric]internal.Gauge{
					"StackSys": 1.0,
				},
				map[internal.Metric]*internal.Queue{
					"Counter": {[]internal.Counter{6}, 1},
				},
			},
			args: args{"StackSys"},
			want: internal.Gauge(1.0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Storage{
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
		dataGauge   map[internal.Metric]internal.Gauge
		dataCounter map[internal.Metric]*internal.Queue
	}
	type args struct {
		metric      internal.Metric
		metricValue any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "LOAD",
			fields: fields{
				map[internal.Metric]internal.Gauge{
					"StackSys": 1.0,
				},
				map[internal.Metric]*internal.Queue{
					"Counter": {[]internal.Counter{6}, 5},
				},
			},
			args: args{"Counter", internal.Counter(6)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Storage{
				DataGauge:   tt.fields.dataGauge,
				DataCounter: tt.fields.dataCounter,
			}
			_ = s.Store(tt.args.metric, tt.args.metricValue)
			require.Equal(t, s.DataCounter["Counter"].GetLength(), internal.Counter(2))
			_ = s.Store(tt.args.metric, tt.args.metricValue)
			require.Equal(t, s.DataCounter["Counter"].GetLength(), internal.Counter(3))
			_ = s.Store(tt.args.metric, tt.args.metricValue)
			require.Equal(t, s.DataCounter["Counter"].GetLength(), internal.Counter(4))
			_ = s.Store(tt.args.metric, tt.args.metricValue)
			require.Equal(t, s.DataCounter["Counter"].GetLength(), internal.Counter(5))
			_ = s.Store(tt.args.metric, tt.args.metricValue)
			require.Equal(t, s.DataCounter["Counter"].GetLength(), internal.Counter(5))
		})
	}
}
