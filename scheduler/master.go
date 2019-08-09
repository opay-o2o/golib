package scheduler

import (
	"context"
	"fmt"
	"github.com/opay-o2o/golib/logger"
)

type Master struct {
	baseCtx  context.Context
	stopFunc context.CancelFunc
	workers  map[string]*Worker
}

func (m *Master) Start() {
	for _, worker := range m.workers {
		go worker.Run()
	}
}

func (m *Master) Stop() {
	m.stopFunc()

	for _, worker := range m.workers {
		worker.Done()
	}
}

func (m *Master) SendSign(provider, signal string) error {
	worker, ok := m.workers[provider]

	if !ok {
		return fmt.Errorf("provider '%s' does not exist", provider)
	}

	return worker.SendSign(signal)
}

func NewMaster(providers []IProvider, l *logger.Logger) *Master {
	master := &Master{workers: make(map[string]*Worker, len(providers))}
	master.baseCtx, master.stopFunc = context.WithCancel(context.Background())

	for _, p := range providers {
		worker, err := NewWorker(p, l, master.baseCtx)

		if err != nil {
			l.Errorf("[%s] init failed | error: %s", p.GetName(), err)
			continue
		}

		master.workers[worker.provider.GetName()] = worker
	}

	return master
}
