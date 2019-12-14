package mqtt

import (
	"errors"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"time"
)

type Config struct {
	Host              string        `toml:"host"`
	Port              int           `toml:"port"`
	Username          string        `toml:"username"`
	Password          string        `toml:"password"`
	QoS               byte          `toml:"qos"`
	ClientId          string        `toml:"client_id"`
	ConnectTimeout    time.Duration `toml:"connect_timeout"`
	DisconnectTimeout uint          `toml:"disconnect_timeout"`
}

func (c *Config) GetAddr() string {
	return fmt.Sprintf("tcp://%s:%d", c.Host, c.Port)
}

func connect(c *Config) (mqtt.Client, error) {
	options := mqtt.NewClientOptions()
	options.AddBroker(c.GetAddr())
	options.SetUsername(c.Username)
	options.SetPassword(c.Password)
	options.SetClientID(c.ClientId)

	client := mqtt.NewClient(options)
	token := client.Connect()

	if ok := token.WaitTimeout(c.ConnectTimeout * time.Millisecond); !ok {
		return nil, errors.New("connection timeout")
	}

	if err := token.Error(); err != nil {
		return nil, err
	}

	return client, nil
}
