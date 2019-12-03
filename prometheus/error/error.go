package grpc

import (
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
)

const (
	VecErrCounterName = "error_code_total"
)

type Prometheus struct {
	counter *prometheus.CounterVec
}

func New(name, env, addr string, vecNames ...string) *Prometheus {
	counterName := VecErrCounterName

	if len(vecNames) > 0 {
		counterName = vecNames[0]
	}

	p := Prometheus{}
	p.counter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        counterName,
			ConstLabels: prometheus.Labels{"service": name, "env": env, "server_addr": addr},
		},
		[]string{"error_code"},
	)
	prometheus.MustRegister(p.counter)

	return &p
}

func (p *Prometheus) Trigger(errorCode int64) {
	p.counter.WithLabelValues(strconv.FormatInt(errorCode, 10)).Inc()
}
