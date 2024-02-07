package server

import (
	// "crypto/tls"
	"fmt"
	// "log"
	"encoding/hex"
	"context"
	"crypto/tls"
	"os"
	"os/signal"
	"syscall"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
	"net"
	"crypto/hmac"
	"github.com/luckyseadog/go-dev/internal/storage"

	"github.com/luckyseadog/go-dev/internal/security"
	"github.com/luckyseadog/go-dev/internal/metrics"


	pb "github.com/luckyseadog/go-dev/protobuf"
)

type MetricsCollectServer struct {
	pb.UnimplementedMetricsCollectServer
	Storage storage.Storage 
}

func (mcs *MetricsCollectServer) AddMetrics(ctx context.Context, in *pb.AddMetricsRequest) (*pb.AddMetricsResponse, error) {
	metricsCurrent := in.Metrics

	

	for _, metric := range metricsCurrent {
		switch metric.MType {
		case "gauge":
			// if metric.Value == -1 || metric.Delta != -1 {
			// 	return nil, status.Error(codes.Unknown, "1Error")
			// }

			if len(in.Key) > 0 {
				computedHash := security.Hash(fmt.Sprintf("%s:gauge:%f", metric.Id, metric.Value), in.Key)
				decodedComputedHash, err := hex.DecodeString(computedHash)
				if err != nil {
					return nil, status.Error(codes.Unknown, "2Error")
				}
				decodedMetricHash, err := hex.DecodeString(metric.Hash)
				if err != nil {
					return nil, status.Error(codes.Unknown, "3Error")
				}
				if !hmac.Equal(decodedComputedHash, decodedMetricHash) {
					return nil, status.Error(codes.Unknown, "4Error")
				}
			}
			err := mcs.Storage.StoreContext(ctx, metrics.Metric(metric.Id), metrics.Gauge(metric.Value))
			if err != nil {
				return nil, status.Error(codes.Unknown, "5Error")
			}

		case "counter":
			// if metric.Delta == -1 || metric.Value != -1 {
			// 	return nil, status.Error(codes.Unknown, "6Error")
			// }

			if len(in.Key) > 0 {
				computedHash := security.Hash(fmt.Sprintf("%s:counter:%d", metric.Id, metric.Delta), in.Key)
				decodedComputedHash, err := hex.DecodeString(computedHash)
				if err != nil {
					return nil, status.Error(codes.Unknown, "7Error")
				}
				decodedMetricHash, err := hex.DecodeString(metric.Hash)
				if err != nil {
					return nil, status.Error(codes.Unknown, "8Error")
				}
				if !hmac.Equal(decodedComputedHash, decodedMetricHash) {
					return nil, status.Error(codes.Unknown, "9Error")
				}
			}
			err := mcs.Storage.StoreContext(ctx, metrics.Metric(metric.Id), metrics.Counter(metric.Delta))
			if err != nil {
				return nil, status.Error(codes.Unknown, "10Error")
			}

		default:
			return nil, status.Error(codes.Unknown, "11Error")
		}
	}

	metricsAnswer := make([]metrics.Metrics, 0)

	for _, metric := range metricsCurrent {
		res := mcs.Storage.LoadContext(ctx, metric.MType, metrics.Metric(metric.Id))
		if res.Err != nil {
			return nil, status.Error(codes.Unknown, "12Error")
		}
		switch metric.MType {
		case "gauge":
			valueFloat64 := float64(res.Value.(metrics.Gauge))
			hashMetric := security.Hash(fmt.Sprintf("%s:gauge:%f", metric.Id, valueFloat64), in.Key)
			metricsAnswer = append(metricsAnswer, metrics.Metrics{ID: metric.Id, MType: metric.MType, Value: &valueFloat64, Hash: hashMetric})
		case "counter":
			valueInt64 := int64(res.Value.(metrics.Counter))
			hashMetric := security.Hash(fmt.Sprintf("%s:counter:%d", metric.Id, valueInt64), in.Key)
			metricsAnswer = append(metricsAnswer, metrics.Metrics{ID: metric.Id, MType: metric.MType, Delta: &valueInt64, Hash: hashMetric})
		default:
			return nil, status.Error(codes.Unknown, "13Error")
		}
	}

	var response pb.AddMetricsResponse

	for _, metric := range metricsAnswer {
		if metric.Delta == nil {
			response.Metrics = append(response.Metrics, &pb.Metric{
				Id: metric.ID,
				MType: metric.MType,
				Value: *metric.Value,
				Hash: metric.Hash,
			})
		}
	
		if metric.Value == nil {
			response.Metrics = append(response.Metrics, &pb.Metric{
				Id: metric.ID,
				MType: metric.MType,
				Delta: *metric.Delta,
				Hash: metric.Hash,
			})
		}
	}

	return &response, nil
}
 

type ServerGRPC struct {
	*grpc.Server
	address string
}

func NewServerGRPC(address string, tlsConfig *tls.Config) *ServerGRPC {
	return &ServerGRPC{
		grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig))), 
		address,
	}
}

func (s *ServerGRPC) Run() {
	fmt.Println("Start server gRPC")
	listen, _ := net.Listen("tcp", s.address)
	// pb.RegisterMetricsCollectServer(s, &MetricsCollectServer{})
	serveChan := make(chan error, 1)
	go func() {
		serveChan <- s.Serve(listen)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case <-stop:
		fmt.Println("shutting down gracefully")
		s.GracefulStop()

	case err := <-serveChan:
		MyLog.Fatal(err)
	}
}