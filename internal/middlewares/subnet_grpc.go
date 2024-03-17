package middlewares

import (
	"context"
	"errors"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func SubnetInterceptor(trustedSubnet string) func(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (any, error) {
		if trustedSubnet == "" {
			resp, err := handler(ctx, req)
			return resp, err
		} else {
			_, ipNet, err := net.ParseCIDR(trustedSubnet)
			if err != nil {
				return nil, err
			}

			md, ok := metadata.FromIncomingContext(ctx)
			if ok {
				values := md.Get("X-Real-IP")
				IP := values[0]
				if !ipNet.Contains(net.ParseIP(IP)) {
					return nil, errors.New("error with SubnetInterceptor")
				}
				resp, err := handler(ctx, req)
				return resp, err

			} else {
				return nil, errors.New("error with SubnetInterceptor")
			}
		}
	}
}
