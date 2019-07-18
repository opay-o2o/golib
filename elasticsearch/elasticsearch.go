package elasticsearch

import (
	"errors"
	"github.com/olivere/elastic/v7"
)

type Config struct {
	Addrs []string `toml:"addrs"`
}

type Client struct {
	config *Config
	client *elastic.Client
}

func (c *Client) start() error {
	client, err := elastic.NewClient(
		elastic.SetURL(c.config.Addrs...),
	)

	if err != nil {
		return err
	}
	c.client = client

	return nil
}

func (c *Client) Get() (*elastic.Client, error) {
	if c.client != nil {
		return c.client, nil
	}

	return nil, errors.New("no elasticsearch client")
}

func (c *Client) Stop() {
	c.client.Stop()
}

func NewElasticSearchClient(c *Config) (*Client, error) {
	client := &Client{
		config: c,
	}

	err := client.start()
	return client, err
}
