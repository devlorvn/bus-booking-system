package metrics

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	// HTTP Metrics
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed.",
		},
		[]string{"path", "method", "status"},
	)

	HttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Latency of HTTP requests in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method", "status"},
	)

	// gRPC Metrics
	GrpcRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests processed.",
		},
		[]string{"service", "method", "status"},
	)

	GrpcRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "Latency of gRPC requests in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "status"},
	)
)

func init() {
	// Register metrics with Prometheus
	prometheus.MustRegister(HttpRequestsTotal)
	prometheus.MustRegister(HttpRequestDuration)
	prometheus.MustRegister(GrpcRequestsTotal)
	prometheus.MustRegister(GrpcRequestDuration)
}

// GinMetricsMiddleware registers Gin HTTP metrics
func GinMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()

		status := strconv.Itoa(c.Writer.Status())
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		method := c.Request.Method

		HttpRequestsTotal.WithLabelValues(path, method, status).Inc()
		HttpRequestDuration.WithLabelValues(path, method, status).Observe(duration)
	}
}

// UnaryServerInterceptor registers gRPC server interceptor metrics
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start).Seconds()

		statusStr := "OK"
		if err != nil {
			if st, ok := status.FromError(err); ok {
				statusStr = st.Code().String()
			} else {
				statusStr = "Unknown"
			}
		}

		parts := strings.Split(info.FullMethod, "/")
		var serviceName, methodName string
		if len(parts) >= 3 {
			serviceName = parts[1]
			methodName = parts[2]
		} else {
			serviceName = "unknown"
			methodName = info.FullMethod
		}

		GrpcRequestsTotal.WithLabelValues(serviceName, methodName, statusStr).Inc()
		GrpcRequestDuration.WithLabelValues(serviceName, methodName, statusStr).Observe(duration)

		return resp, err
	}
}

// StartMetricsServer starts a background HTTP server for scraping metrics
func StartMetricsServer(port string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		slog.Info("Starting metrics server", slog.String("port", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Metrics server failed to start", slog.String("error", err.Error()))
		}
	}()
}
