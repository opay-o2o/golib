package grpc

import (
	"context"
	"fmt"
	"github.com/opay-o2o/golib/strings2"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/peer"
	"net"
	"strings"
	"time"
)

var DefaultBuckets = []float64{0.1, 0.3, 0.5, 1.0, 3.0, 5.0}

const (
	VecReqCounterName  = "grpc_requests_total"
	VecReqLatencyName  = "grpc_request_duration_seconds"
	VecCallCounterName = "grpc_call_total"
	VecCallLatencyName = "grpc_call_duration_seconds"
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

type Prometheus struct {
	counter *prometheus.CounterVec
	latency *prometheus.HistogramVec
}

func New(name, env string, vecNames ...string) *Prometheus {
	counterName := strings2.IIf(len(vecNames) > 0, vecNames[0], VecReqCounterName)
	latencyName := strings2.IIf(len(vecNames) > 1, vecNames[1], VecReqLatencyName)

	p := Prometheus{}
	p.counter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        counterName,
			ConstLabels: prometheus.Labels{"service": name, "env": env},
		},
		[]string{"method", "server_addr", "client_ip"},
	)
	prometheus.MustRegister(p.counter)

	p.latency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        latencyName,
		ConstLabels: prometheus.Labels{"service": name, "env": env},
		Buckets:     DefaultBuckets,
	},
		[]string{"method", "server_addr", "client_ip"},
	)
	prometheus.MustRegister(p.latency)

	return &p
}

func (p *Prometheus) Trigger(ctx context.Context, addr, method string, startTime time.Time) {
	clientIp, err := getClietIP(ctx)

	if err != nil {
		clientIp = "unknown"
	}

	p.counter.WithLabelValues(method, addr, clientIp).Inc()

	useTime := float64(time.Since(startTime).Nanoseconds()) / 1000000000
	p.latency.WithLabelValues(method, addr, clientIp).Observe(useTime)
}
