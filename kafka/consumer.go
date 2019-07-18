package kafka

import (
	"../logger"
	"context"
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"sync"
)

type ConsumerConfig struct {
	Brokers      []string `toml:"brokers"`
	Topic        string   `toml:"topic"`
	Group        string   `toml:"group"`
	Worker       int      `toml:"worker"`
	OffsetNewest bool     `toml:"offset_newest"`
}

type Consumer struct {
	c       *ConsumerConfig
	client  *cluster.Consumer
	handler func([]byte) error
	logger  *logger.Logger
	ctx     context.Context
	cancel  context.CancelFunc
	wg      *sync.WaitGroup
}

func (c *Consumer) run() {
	c.wg.Add(c.c.Worker + 2)

	for i := 0; i < c.c.Worker; i++ {
		go c.receive()
	}

	go c.logErr()
	go c.logNotice()
}

func (c *Consumer) Stop() {
	c.cancel()

	if err := c.client.Close(); err != nil {
		c.logger.Errorf("kafka consumer close failed | brokers: %+v | group: %s | error: %s", c.c.Brokers, c.c.Group, err)
	}

	c.wg.Wait()
}

func (c *Consumer) logErr() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case err := <-c.client.Errors():
			c.logger.Errorf("kafka consumer revice error | brokers: %+v | group: %s | error: %s", c.c.Brokers, c.c.Group, err)
		}
	}
}

func (c *Consumer) logNotice() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.client.Notifications():
			c.logger.Debugf("kafka consumer revice notification | brokers: %+v | group: %s | message: %s", c.c.Brokers, c.c.Group, msg)
		}
	}
}

func (c *Consumer) receive() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.client.Messages():
			err := c.handler(msg.Value)

			if err != nil {
				c.logger.Errorf("kafka consumer handler error | brokers: %+v | group: %s | error: %s", c.c.Brokers, c.c.Group, err)
				continue
			}

			c.client.MarkOffset(msg, "done")
		}
	}
}

func NewConsumer(c *ConsumerConfig, handler func([]byte) error, logger *logger.Logger) (consumer *Consumer, err error) {
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true

	if c.OffsetNewest {
		config.Consumer.Offsets.Initial = sarama.OffsetNewest
	} else {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	client, err := cluster.NewConsumer(c.Brokers, c.Group, []string{c.Topic}, config)

	if err != nil {
		return
	}

	consumer = &Consumer{
		c:       c,
		client:  client,
		handler: handler,
		logger:  logger,
		wg:      &sync.WaitGroup{},
	}

	consumer.ctx, consumer.cancel = context.WithCancel(context.Background())
	consumer.run()
	return
}
