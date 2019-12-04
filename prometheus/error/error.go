package grpc

import (
	"github.com/opay-o2o/golib/strings2"
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
		[]string{"error_code", "critical"},
	)
	prometheus.MustRegister(p.counter)

	return &p
}

func (p *Prometheus) Trigger(errorCode int64, critical bool) {
	p.counter.WithLabelValues(strconv.FormatInt(errorCode, 10), strings2.IIf(critical, "1", "0")).Inc()
}
