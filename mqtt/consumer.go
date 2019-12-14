package mqtt

import (
	"context"
	"errors"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/opay-o2o/golib/logger"
	"sync"
	"time"
)

type ConsumerConfig struct {
	*Config
	Timeout time.Duration `toml:"timeout"`
	Worker  int           `toml:"worker"`
}

type Consumer struct {
	c        *ConsumerConfig
	client   mqtt.Client
	msgQueue chan mqtt.Message
	handlers map[string]func([]byte)
	logger   *logger.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	wg       *sync.WaitGroup
}

func (c *Consumer) run() (err error) {
	if c.client, err = connect(c.c.Config); err != nil {
		return
	}

	filters := make(map[string]byte, len(c.handlers))

	for t := range c.handlers {
		filters[t] = c.c.QoS
	}

	token := c.client.SubscribeMultiple(filters, func(client mqtt.Client, msg mqtt.Message) {
		c.msgQueue <- msg
	})

	if ok := token.WaitTimeout(c.c.Timeout * time.Millisecond); !ok {
		return errors.New("subscribe timeout")
	}

	if err = token.Error(); err != nil {
		return
	}

	c.wg.Add(c.c.Worker)

	for i := 0; i < c.c.Worker; i++ {
		go c.handle()
	}

	return
}

func (c *Consumer) Stop() {
	topics := make([]string, 0, len(c.handlers))

	for t := range c.handlers {
		topics = append(topics, t)
	}

	token := c.client.Unsubscribe(topics...)

	if ok := token.WaitTimeout(c.c.Timeout * time.Millisecond); !ok {
		c.logger.Errorf("unsubscribe topics timeout | addr: %s | topics: %+v", c.c.GetAddr(), topics)
	}

	if err := token.Error(); err != nil {
		c.logger.Errorf("can't unsubscribe topics | addr: %s | topics: %+v | error: %s", c.c.GetAddr(), topics, err)
	}

	c.client.Disconnect(c.c.DisconnectTimeout)

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
			if handler, ok := c.handlers[msg.Topic()]; ok {
				handler(msg.Payload())
			}
		}
	}
}

func NewConsumer(c *ConsumerConfig, handlers map[string]func([]byte), logger *logger.Logger) (consumer *Consumer, err error) {
	consumer = &Consumer{
		c:        c,
		msgQueue: make(chan mqtt.Message, 4096),
		handlers: handlers,
		logger:   logger,
		wg:       &sync.WaitGroup{},
	}

	consumer.ctx, consumer.cancel = context.WithCancel(context.Background())
	err = consumer.run()
	return
}
