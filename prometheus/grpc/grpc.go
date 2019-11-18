package grpc

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/peer"
	"net"
	"strings"
	"time"
)

var (
	// DefaultBuckets prometheus buckets in seconds.
	DefaultBuckets = []float64{0.1, 0.3, 0.5, 1.0, 3.0, 5.0}
)

const (
	reqsName    = "grpc_requests_total"
	latencyName = "grpc_request_duration_seconds"
)

func getClietIP(ctx context.Context) (string, error) {
	pr, ok := peer.FromContext(ctx)

	if !ok {
		return "", fmt.Errorf("invoke FromContext() failed")
	}

	if pr.Addr == net.Addr(nil) {
		return "", fmt.Errorf("peer.Addr is nil")
	}

	return strings.Split(pr.Addr.String(), ":")[0], nil
}

// Prometheus is a handler that exposes prometheus metrics for the number of requests,
// the latency and the response size, partitioned by status code, method and HTTP path.
//
// Usage: pass its `ServeHTTP` to a route or globally.
type Prometheus struct {
	reqs    *prometheus.CounterVec
	latency *prometheus.HistogramVec
}

// New returns a new prometheus middleware.
//
// If buckets are empty then `DefaultBuckets` are set.
func New(name, env, addr string, buckets ...float64) *Prometheus {
	p := Prometheus{}
	p.reqs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        reqsName,
			Help:        "How many GRPC requests processed, partitioned by method, server ip, client ip.",
			ConstLabels: prometheus.Labels{"service": name, "env": env, "server_addr": addr},
		},
		[]string{"method", "client_ip"},
	)
	prometheus.MustRegister(p.reqs)

	if len(buckets) == 0 {
		buckets = DefaultBuckets
	}

	p.latency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        latencyName,
		Help:        "How long it took to process the request, partitioned by method, server ip, client ip.",
		ConstLabels: prometheus.Labels{"service": name, "env": env, "server_addr": addr},
		Buckets:     buckets,
	},
		[]string{"method", "client_ip"},
	)
	prometheus.MustRegister(p.latency)

	return &p
}

func (p *Prometheus) Trigger(ctx context.Context, method string, startTime time.Time) {
	clientIp, err := getClietIP(ctx)

	if err != nil {
		clientIp = "unknown"
	}

	p.reqs.WithLabelValues(method, clientIp).Inc()

	useTime := float64(time.Since(startTime).Nanoseconds()) / 1000000000
	p.latency.WithLabelValues(method, clientIp).Observe(useTime)
}
