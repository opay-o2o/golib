package scheduler

import (
	"context"
	"fmt"
	"github.com/opay-o2o/golib/logger"
	"strings"
	"time"
)

type Point struct {
	signal bool
	t      time.Time
}

type Worker struct {
	provider  IProvider
	logger    *logger.Logger
	ctx       context.Context
	loopTimer chan *Point
	endSign   chan bool
}

func (w *Worker) SendSign(signal string) error {
	select {
	case <-w.ctx.Done():
		return fmt.Errorf("[%s] worker is closed", w.provider.GetName())
	default:
		signal := strings.TrimSpace(signal)

		if len(signal) == 0 {
			return fmt.Errorf("[%s] signal fotmat error: '%s'", w.provider.GetName(), signal)
		}

		signTime, err := time.ParseInLocation(SignalFormat, signal, time.Local)

		if err != nil {
			return fmt.Errorf("[%s] signal fotmat error: '%s'", w.provider.GetName(), signal)
		}

		w.loopTimer <- &Point{true, signTime}
	}

	return nil
}

func (w *Worker) startLoop() {
	if second := time.Now().Second(); second > 0 {
		time.Sleep(time.Minute - time.Duration(second)*time.Second)
	}

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case t := <-ticker.C:
			w.loopTimer <- &Point{false, t}
		}
	}
}

func (w *Worker) Run() {
	defer func() {
		w.logger.Infof("[%s] worker end", w.provider.GetName())
		w.endSign <- true
	}()

	w.logger.Infof("[%s] worker start", w.provider.GetName())
	go w.startLoop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case p := <-w.loopTimer:
			if p.signal {
				w.logger.Infof("[%s] run by signal file", w.provider.GetName())
				w.provider.Run(p.t)
			} else if w.provider.CheckInterval(p.t) {
				w.provider.Run(p.t)
			}
		}
	}
}

func (w *Worker) Done() {
	<-w.endSign
}

func NewWorker(p IProvider, l *logger.Logger, ctx context.Context) (worker *Worker, err error) {
	if err = p.Init(); err != nil {
		return
	}

	worker = &Worker{
		provider:  p,
		logger:    l,
		ctx:       ctx,
		loopTimer: make(chan *Point, 1),
		endSign:   make(chan bool, 1),
	}
	return
}
