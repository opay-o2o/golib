package mqtt

import (
	"context"
	"errors"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/opay-o2o/golib/logger"
	"sync"
	"time"
)

type Message struct {
	Topic    string
	Retained bool
	Payload  []byte
}

func (m *Message) String() string {
	return fmt.Sprintf("{Topic:%s Retained:%+v Payload:%s}", m.Topic, m.Retained, m.Payload)
}

type ProducerConfig struct {
	*Config
	Timeout         time.Duration `toml:"timeout"`
	RetryInterval   time.Duration `toml:"retry_interval"`
	ReconnectOnFail int           `toml:"reconnect_on_fail"`
}

type Producer struct {
	c        *ProducerConfig
	client   mqtt.Client
	msgQueue chan *Message
	logger   *logger.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	wg       *sync.WaitGroup
}

func (c *Producer) run() (err error) {
	if c.client, err = connect(c.c.Config); err != nil {
		return
	}

	c.wg.Add(1)
	go c.publish()

	return
}

func (c *Producer) Stop() {
	c.cancel()
	c.wg.Wait()
	c.client.Disconnect(c.c.DisconnectTimeout)
}

func (c *Producer) Send(topic string, retained bool, payload []byte) error {
	select {
	case <-c.ctx.Done():
		return errors.New("producer is stoped")
	default:
		c.msgQueue <- &Message{topic, retained, payload}
		return nil
	}
}

func (c *Producer) retry(msg *Message, times *int) {
	c.msgQueue <- msg
	*times++

	time.Sleep(c.c.RetryInterval * time.Millisecond)

	if *times >= c.c.ReconnectOnFail {
		client, err := connect(c.c.Config)

		if err != nil {
			c.logger.Errorf("can't connect to server | addr: %s | error: %s", c.c.GetAddr(), err)
			return
		}

		c.client, *times = client, 0
	}
}

func (c *Producer) publish() {
	defer c.wg.Done()

	retryTimes := 0

	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.msgQueue:
			token := c.client.Publish(msg.Topic, c.c.QoS, msg.Retained, msg.Payload)

			if ok := token.WaitTimeout(c.c.Timeout * time.Millisecond); !ok {
				c.logger.Errorf("publish message timeout | addr: %s | message: %s", c.c.GetAddr(), msg)
				c.retry(msg, &retryTimes)
				continue
			}

			if err := token.Error(); err != nil {
				c.logger.Errorf("can't publish message | addr: %s | message: %s | error: %s", c.c.GetAddr(), msg, err)
				c.retry(msg, &retryTimes)
				continue
			}

			c.logger.Debugf("publish message | addr: %s | message: %s", c.c.GetAddr(), msg)
		}
	}
}

func NewProducer(c *ProducerConfig, logger *logger.Logger) (producer *Producer, err error) {
	producer = &Producer{
		c:        c,
		msgQueue: make(chan *Message, 4096),
		logger:   logger,
		wg:       &sync.WaitGroup{},
	}

	producer.ctx, producer.cancel = context.WithCancel(context.Background())
	err = producer.run()
	return
}
