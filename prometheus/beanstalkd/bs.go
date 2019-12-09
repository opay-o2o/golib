package beanstalkd

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	VecGaugeName = "beanstalkd_stats"
)

type Prometheus struct {
	gauger *prometheus.GaugeVec
}

func New(name, env, addr string) *Prometheus {
	p := Prometheus{}
	p.gauger = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        VecGaugeName,
			ConstLabels: prometheus.Labels{"service": name, "env": env, "server_addr": addr},
		},
		[]string{"tube", "key"},
	)
	prometheus.MustRegister(p.gauger)

	return &p
}

func (p *Prometheus) Trigger(tube, key string, value int64) {
	p.gauger.WithLabelValues(tube, key).Set(float64(value))
}
