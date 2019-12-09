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

func New(name, env string) *Prometheus {
	p := Prometheus{}
	p.gauger = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        VecGaugeName,
			ConstLabels: prometheus.Labels{"service": name, "env": env},
		},
		[]string{"addr", "tube", "key"},
	)
	prometheus.MustRegister(p.gauger)

	return &p
}

func (p *Prometheus) Trigger(addr, tube, key string, value int64) {
	p.gauger.WithLabelValues(addr, tube, key).Set(float64(value))
}
