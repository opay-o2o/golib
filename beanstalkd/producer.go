package beanstalkd

import (
	"context"
	"errors"
	"fmt"
	"github.com/kr/beanstalk"
	"github.com/opay-o2o/golib/logger"
	"sync"
	"time"
)

type ProducerConfig struct {
	Addrs []string `toml:"addrs"`
}

type Message struct {
	Tube     string
	Payload  []byte
	Priority uint32
	Delay    time.Duration
	Ttr      time.Duration
}

func (m *Message) String() string {
	return fmt.Sprintf("{Tube:%s Payload:%s Priority:%d Delay:%s Ttr:%s}", m.Tube, m.Payload, m.Priority, m.Delay, m.Ttr)
}

type Producer struct {
	c        *ProducerConfig
	msgQueue chan *Message
	logger   *logger.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	wg       *sync.WaitGroup
}

func (c *Producer) run() {
	c.wg.Add(len(c.c.Addrs))

	for _, addr := range c.c.Addrs {
		go c.send(addr)
	}
}

func (c *Producer) Stop() {
	c.cancel()
	c.wg.Wait()
}

func (c *Producer) Send(msg *Message) error {
	select {
	case <-c.ctx.Done():
		return errors.New("producer is stoped")
	default:
		c.msgQueue <- msg
		return nil
	}
}

func (c *Producer) send(addr string) {
	defer c.wg.Done()

	var (
		err  error
		tube *beanstalk.Tube
	)

	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.msgQueue:
			if tube == nil {
				tube, err = newTube(addr, msg.Tube)

				if err != nil {
					c.logger.Errorf("can't connect beanstalkd server | addr: %s | tube: %s | error: %s", addr, msg.Tube, err)

					tube = nil
					time.Sleep(time.Second)
					c.msgQueue <- msg

					continue
				}
			} else {
				tube.Name = msg.Tube
			}

			id, err := tube.Put(msg.Payload, msg.Priority, msg.Delay, msg.Ttr)

			if err != nil {
				c.logger.Errorf("can't create job | addr: %s | msg: %s | error: %s", addr, msg, err)

				tube = nil
				time.Sleep(3 * time.Second)
				c.msgQueue <- msg

				continue
			}

			c.logger.Debugf("create job | id: %d | msg: %s", id, msg)
		}
	}
}

func newTube(addr string, tube string) (*beanstalk.Tube, error) {
	conn, err := beanstalk.Dial("tcp", addr)

	if err != nil {
		return nil, err
	}

	return &beanstalk.Tube{Conn: conn, Name: tube}, nil
}

func NewProducer(c *ProducerConfig, logger *logger.Logger) *Producer {
	producer := &Producer{
		c:        c,
		msgQueue: make(chan *Message, 32),
		logger:   logger,
		wg:       &sync.WaitGroup{},
	}

	producer.ctx, producer.cancel = context.WithCancel(context.Background())
	producer.run()
	return producer
}
