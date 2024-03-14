package agent

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
	"path"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/luckyseadog/go-dev/protobuf"
)

type AgentGRPC struct {
	client  pb.MetricsCollectClient
	ruler   InteractionRules
	metrics Metrics
	mu      sync.RWMutex
	cancel  chan struct{}
}

func NewAgentGRPC(address string, contentType string, pollInterval time.Duration, reportInterval time.Duration, secretKey []byte, rateLimit int, cryptoKeyDir string) (*AgentGRPC, error) {
	rateLimitChan := make(chan struct{}, rateLimit)
	for i := 0; i < rateLimit; i++ {
		rateLimitChan <- struct{}{}
	}
	interactionRules := InteractionRules{
		address:        address,
		contentType:    contentType,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		secretKey:      secretKey,
		rateLimitChan:  rateLimitChan,
	}
	cancel := make(chan struct{})

	if cryptoKeyDir != "" {
		clientTLSCert, err := tls.LoadX509KeyPair(path.Join(cryptoKeyDir, "agent/certAgent.pem"), path.Join(cryptoKeyDir, "agent/privateKeyAgent.pem"))
		if err != nil {
			return nil, err
		}

		// Configure the client to trust TLS server certs issued by a CA.
		certPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		if caCertPEM, err := os.ReadFile(path.Join(cryptoKeyDir, "root/certRoot.pem")); err != nil {
			return nil, err
		} else if ok := certPool.AppendCertsFromPEM(caCertPEM); !ok {
			return nil, errors.New("invalid cert in CA PEM")
		}
		tlsConfig := &tls.Config{
			RootCAs:      certPool,
			Certificates: []tls.Certificate{clientTLSCert},
		}

		c, err := grpc.Dial(
			address,
			grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		)
		if err != nil {
			return nil, err
		}
		return &AgentGRPC{client: pb.NewMetricsCollectClient(c), ruler: interactionRules, cancel: cancel}, nil
	} else {
		c, err := grpc.Dial(
			address,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, err
		}
		return &AgentGRPC{client: pb.NewMetricsCollectClient(c), ruler: interactionRules, cancel: cancel}, nil
	}
}
