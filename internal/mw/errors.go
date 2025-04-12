package mw

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"pvz/internal/logger"
)

func ErrorInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		span, ctx := opentracing.StartSpanFromContext(ctx, "ErrorInterceptor")
		defer span.Finish()

		resp, err = handler(ctx, req)
		if err != nil {
			logger.Error(err)
		}
		return resp, err
	}
}
