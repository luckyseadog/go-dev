package middlewares

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func GzipInterceptor(ctx context.Context, req any,
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		values := md.Get("Content-Encoding")
		for _, value := range values {
			if strings.ToLower(value) == "gzip" {
				if compressedReq, ok := req.([]byte); ok {
					uncompressedReq, err := decompressGzip(compressedReq)
					if err != nil {
						return nil, err
					}
					req = uncompressedReq
				} else {
					return nil, errors.New("error with gzipInterceptor")
				}
			}
		}
	}

	resp, err := handler(ctx, req)
	if err != nil {
		return resp, err
	}

	md, ok = metadata.FromIncomingContext(ctx)
	if ok {
		values := md.Get("Accept-Encoding")
		for _, value := range values {
			if strings.ToLower(value) == "gzip" {
				if uncompressedReply, ok := resp.([]byte); ok {
					compressedReply, err := compressGzip(uncompressedReply)
					if err != nil {
						return nil, err
					}
					resp = compressedReply
				} else {
					return nil, errors.New("error with GzipInterceptor")
				}
			}
		}
	}

	return resp, nil
}

func decompressGzip(input []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(input))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

func compressGzip(data []byte) ([]byte, error) {
	var compressedBuffer bytes.Buffer
	writer := gzip.NewWriter(&compressedBuffer)

	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return compressedBuffer.Bytes(), nil
}
