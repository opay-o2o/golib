package prometheus

import (
	stdCtx "context"
	"errors"
	"fmt"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"github.com/opay-o2o/golib/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc/peer"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	TypeCounter = iota
	TypeHistogram
	TypeGauge
	TypeSummary
)

var (
	// DefaultBuckets prometheus buckets in seconds.
	DefaultBuckets = []float64{0.3, 1.2, 5.0}
)

type VectorConfig struct {
	Name   string   `toml:"name"`
	Desc   string   `toml:"desc"`
	Type   int      `toml:"type"`
	Labels []string `toml:"labels"`
}

type Config struct {
	Env     string          `toml:"env"`
	Service string          `toml:"service"`
	Host    string          `toml:"host"`
	Vectors []*VectorConfig `toml:"vectors"`
}

type Vector struct {
	config *VectorConfig
	vec    prometheus.Collector
}

func (v *Vector) Trigger(value float64, labels ...string) {
	switch v.config.Type {
	case TypeHistogram:
		v.vec.(*prometheus.HistogramVec).WithLabelValues(labels...).Observe(value)
	case TypeGauge:
		v.vec.(*prometheus.GaugeVec).WithLabelValues(labels...).Set(value)
	case TypeSummary:
		v.vec.(*prometheus.SummaryVec).WithLabelValues(labels...).Observe(value)
	case TypeCounter:
		v.vec.(*prometheus.CounterVec).WithLabelValues(labels...).Inc()
	}
}

type Monitor struct {
	config  *Config
	vectors map[string]*Vector
	logger  *logger.Logger
}

func (m *Monitor) Register(config *VectorConfig) (err error) {
	var vec prometheus.Collector
	constLabels := map[string]string{"service": m.config.Service, "env": m.config.Env, "host": m.config.Host}

	switch config.Type {
	case TypeHistogram:
		vec = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        config.Name,
				Help:        config.Desc,
				ConstLabels: constLabels,
				Buckets:     DefaultBuckets,
			},
			config.Labels,
		)
	case TypeGauge:
		vec = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:        config.Name,
				Help:        config.Desc,
				ConstLabels: constLabels,
			},
			config.Labels,
		)
	case TypeSummary:
		vec = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:        config.Name,
				Help:        config.Desc,
				ConstLabels: constLabels,
			},
			config.Labels,
		)
	case TypeCounter:
		vec = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        config.Name,
				Help:        config.Desc,
				ConstLabels: constLabels,
			},
			config.Labels,
		)
	default:
		err = errors.New("invalid monitor type")
		return
	}

	m.vectors[config.Name] = &Vector{config: config, vec: vec}
	prometheus.MustRegister(vec)
	return
}

func (m *Monitor) Trigger(name string, value float64, labels ...string) {
	if vector, ok := m.vectors[name]; ok {
		vector.Trigger(value, labels...)
	} else {
		m.logger.Warningf("unknown monitor vector '%s'", name)
	}
}

func (m *Monitor) Vector(name string) (vector *Vector) {
	return m.vectors[name]
}

func (m *Monitor) Group(counterName, timerName string) (group *VectorGroup, err error) {
	counterVector, ok := m.vectors[counterName]

	if !ok {
		err = errors.New("unknown monitor vector")
		return
	}

	timerVector, ok := m.vectors[timerName]

	if !ok {
		err = errors.New("unknown monitor vector")
		return
	}

	group = &VectorGroup{counterVector, timerVector, m.logger}
	return
}

type VectorGroup struct {
	counter *Vector
	timer   *Vector
	logger  *logger.Logger
}

func (g *VectorGroup) HttpTrigger(ctx context.Context) {
	start := time.Now()
	ctx.Next()
	r := ctx.Request()
	statusCode := strconv.Itoa(ctx.GetStatusCode())
	duration := float64(time.Since(start).Nanoseconds()) / 1000000000

	g.counter.Trigger(0, statusCode, r.Method, r.URL.Path)
	g.timer.Trigger(duration, statusCode, r.Method, r.URL.Path)
}

func getClietIP(ctx stdCtx.Context) (string, error) {
	pr, ok := peer.FromContext(ctx)

	if !ok {
		return "", fmt.Errorf("invoke FromContext() failed")
	}

	if pr.Addr == net.Addr(nil) {
		return "", fmt.Errorf("peer.Addr is nil")
	}

	return strings.Split(pr.Addr.String(), ":")[0], nil
}

func (g *VectorGroup) GrpcTrigger(ctx stdCtx.Context, addr, method string, startTime time.Time) {
	clientIp, err := getClietIP(ctx)

	if err != nil {
		clientIp = "unknown"
	}

	duration := float64(time.Since(startTime).Nanoseconds()) / 1000000000

	g.counter.Trigger(0, method, addr, clientIp)
	g.timer.Trigger(duration, method, addr, clientIp)
}

func (m *Monitor) Metrics() context.Handler {
	return iris.FromStd(promhttp.Handler())
}

func NewMonitor(config *Config, logger *logger.Logger) (monitor *Monitor, err error) {
	monitor = &Monitor{
		config:  config,
		vectors: make(map[string]*Vector, len(config.Vectors)),
		logger:  logger,
	}

	for _, c := range config.Vectors {
		if err = monitor.Register(c); err != nil {
			break
		}
	}

	return
}
