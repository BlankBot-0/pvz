package mw

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"time"
)

var handledRequests = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "handled_requests",
	Help:    "Duration of handling requests with distribution by method/status.",
	Buckets: responseTimeBuckets,
}, []string{"method", "status"})

func observeHandledRequest(duration time.Duration, method, status string) {
	handledRequests.WithLabelValues(method, status).Observe(duration.Seconds())
}

func MetricInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		span, ctx := opentracing.StartSpanFromContext(ctx, "metricInterceptor")
		defer span.Finish()

		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		stat, _ := status.FromError(err)
		observeHandledRequest(duration, info.FullMethod, stat.Code().String())
		return resp, err
	}
}

var (
	responseTimeBuckets = []float64{0.001, 0.003, 0.007, 0.015, 0.05, 0.1, 0.2, 0.5, 1, 2, 5}
)
