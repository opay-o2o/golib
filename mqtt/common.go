package mqtt

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"time"
)

type Config struct {
	Host              string `toml:"host"`
	Port              int    `toml:"port"`
	Username          string `toml:"username"`
	Password          string `toml:"password"`
	QoS               byte   `toml:"qos"`
	ClientName        string `toml:"client_name"`
	ConnectTimeout    uint   `toml:"connect_timeout"`
	DisconnectTimeout uint   `toml:"disconnect_timeout"`
}

func (c *Config) GetAddr() string {
	return fmt.Sprintf("tcp://%s:%d", c.Host, c.Port)
}

func connect(c *Config) (mqtt.Client, error) {
	options := mqtt.NewClientOptions()
	options.AddBroker(c.GetAddr())
	options.SetUsername(c.Username)
	options.SetPassword(c.Password)
	options.SetClientID(c.ClientName)

	client := mqtt.NewClient(options)
	token := client.Connect()
	token.WaitTimeout(time.Duration(c.ConnectTimeout) * time.Millisecond)

	if err := token.Error(); err != nil {
		return nil, err
	}

	return client, nil
}
