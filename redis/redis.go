package redis

import (
	"errors"
	"github.com/go-redis/redis"
	"strconv"
	"sync"
)

type Config struct {
	Host     string `toml:"host" json:"host"`
	Port     int    `toml:"port" json:"port"`
	Password string `toml:"password" json:"password"`
	Database int    `toml:"database" json:"database"`
	PoolSize int    `toml:"poolsize" json:"poolsize"`
}

func (c *Config) GetAddr() string {
	return c.Host + ":" + strconv.Itoa(c.Port)
}

type Pool struct {
	locker  sync.RWMutex
	clients map[string]*redis.Client
}

func (p *Pool) Add(name string, c *Config) {
	p.locker.Lock()
	defer p.locker.Unlock()

	p.clients[name] = redis.NewClient(&redis.Options{Addr: c.GetAddr(), Password: c.Password, DB: c.Database, PoolSize: c.PoolSize})
}

func (p *Pool) Get(name string) (*redis.Client, error) {
	p.locker.RLock()
	defer p.locker.RUnlock()

	client, ok := p.clients[name]

	if ok {
		return client, nil
	}

	return nil, errors.New("no redis client")
}

func NewPool() *Pool {
	return &Pool{clients: make(map[string]*redis.Client, 64)}
}
