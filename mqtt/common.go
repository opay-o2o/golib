package mqtt

import (
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
	ReceiveSaveMsg    bool          `toml:"receive_save_msg"`
	ConnectTimeout    time.Duration `toml:"connect_timeout"`
	DisconnectTimeout uint          `toml:"disconnect_timeout"`
}

func (c *Config) GetAddr() string {
	return fmt.Sprintf("tcp://%s:%d", c.Host, c.Port)
}

func connect(c *Config) (mqtt.Client, error) {
	options := mqtt.NewClientOptions().
		AddBroker(c.GetAddr()).
		SetUsername(c.Username).
		SetPassword(c.Password).
		SetClientID(c.ClientId).
		SetCleanSession(!c.ReceiveSaveMsg).
		SetConnectTimeout(c.ConnectTimeout * time.Millisecond)
	client := mqtt.NewClient(options)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return client, nil
}
