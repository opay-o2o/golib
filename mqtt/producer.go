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
	Timeout uint `toml:"timeout"`
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
	go c.send()

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

func (c *Producer) send() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.msgQueue:
			token := c.client.Publish(msg.Topic, c.c.QoS, msg.Retained, msg.Payload)
			token.WaitTimeout(time.Duration(c.c.Timeout) * time.Millisecond)

			if err := token.Error(); err != nil {
				c.logger.Errorf("can't publish message | addr: %s | message: %s | error: %s", c.c.GetAddr(), msg, err)
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
