package beanstalkd

import (
	"../logger"
	"context"
	"github.com/kr/beanstalk"
	"sync"
	"time"
)

type AddrList struct {
	Addrs []string `toml:"addrs"`
}

type ConsumerConfig struct {
	Addrs  []string `toml:"addrs"`
	Tube   string   `toml:"tube"`
	Worker int      `toml:"worker"`
}

type Consumer struct {
	c        *ConsumerConfig
	msgQueue chan []byte
	handler  func([]byte)
	logger   *logger.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	wg       *sync.WaitGroup
}

func (c *Consumer) run() {
	c.wg.Add(len(c.c.Addrs) + c.c.Worker)

	for _, addr := range c.c.Addrs {
		go c.receive(addr)
	}

	for i := 0; i < c.c.Worker; i++ {
		go c.handle()
	}
}

func (c *Consumer) Stop() {
	c.cancel()
	c.wg.Wait()
}

func (c *Consumer) handle() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.msgQueue:
			c.handler(msg)
		}
	}
}

func (c *Consumer) receive(addr string) {
	defer c.wg.Done()

	var (
		tubeSet *beanstalk.TubeSet
		err     error
	)

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			if tubeSet == nil {
				tubeSet, err = newTubeSet(addr, c.c.Tube)

				if err != nil {
					c.logger.Errorf("can't connect beanstalkd server | addr: %s | tube: %s | error: %s", addr, c.c.Tube, err)
					time.Sleep(time.Second)
					continue
				}
			}

			id, body, err := tubeSet.Reserve(5 * time.Second)

			if err != nil {
				if e, ok := err.(beanstalk.ConnError); ok && e.Err == beanstalk.ErrTimeout {
					continue
				}

				c.logger.Errorf("can't reserve job | addr: %s | tube: %s | error: %s", addr, c.c.Tube, err)

				tubeSet = nil
				time.Sleep(3 * time.Second)
				continue
			}

			c.msgQueue <- body

			if err := tubeSet.Conn.Delete(id); err != nil {
				c.logger.Errorf("can't delete job | addr: %s | tube: %s | id: %d | error: %s", addr, c.c.Tube, id, err)
			}
		}
	}

}

func newTubeSet(addr string, tube string) (*beanstalk.TubeSet, error) {
	conn, err := beanstalk.Dial("tcp", addr)

	if err != nil {
		return nil, err
	}

	return beanstalk.NewTubeSet(conn, tube), nil
}

func NewConsumer(c *ConsumerConfig, handler func([]byte), logger *logger.Logger) *Consumer {
	consumer := &Consumer{
		c:        c,
		msgQueue: make(chan []byte, 32),
		handler:  handler,
		logger:   logger,
		wg:       &sync.WaitGroup{},
	}

	consumer.ctx, consumer.cancel = context.WithCancel(context.Background())
	consumer.run()
	return consumer
}
