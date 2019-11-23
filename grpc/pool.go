package grpc

import (
	"google.golang.org/grpc"
	"sync"
	"time"
)

type Dialer func(addr string) (*grpc.ClientConn, error)

func DefaultDialer(addr string) (*grpc.ClientConn, error) {
	return grpc.Dial(addr, grpc.WithInsecure())
}

type Pool struct {
	sync.RWMutex
	conns    map[string]chan *Connection
	dialer   Dialer
	capacity int
	idle     time.Duration
	ttl      time.Duration
}

type Connection struct {
	addr     string
	conn     *grpc.ClientConn
	pool     *Pool
	createAt time.Time
	lastUsed time.Time
}

func NewPool(dialer Dialer, capacity int, idle time.Duration, ttl ...time.Duration) *Pool {
	if capacity <= 0 {
		capacity = 1
	}

	p := &Pool{
		conns:    make(map[string]chan *Connection, 16),
		capacity: capacity,
		dialer:   dialer,
		idle:     idle,
	}

	if len(ttl) > 0 {
		p.ttl = ttl[0]
	}

	return p
}

func (p *Pool) Get(addr string) (*Connection, error) {
	p.Lock()
	clients, ok := p.conns[addr]

	if !ok {
		clients = make(chan *Connection, p.capacity)
		p.conns[addr] = clients
	}

	p.Unlock()

	for {
		select {
		case client := <-clients:
			if p.idle > 0 && client.lastUsed.Add(p.idle).Before(time.Now()) {
				_ = client.conn.Close()
				continue
			}

			if p.ttl > 0 && client.createAt.Add(p.ttl).Before(time.Now()) {
				_ = client.conn.Close()
				continue
			}

			client.lastUsed = time.Now()
			return client, nil
		default:
			c, err := p.dialer(addr)

			if err != nil {
				return nil, err
			}

			client := &Connection{
				addr:     addr,
				pool:     p,
				conn:     c,
				createAt: time.Now(),
				lastUsed: time.Now(),
			}

			return client, nil
		}
	}
}

func (c *Connection) GetConn() *grpc.ClientConn {
	return c.conn
}

func (c *Connection) Close() {
	if c == nil || c.conn == nil {
		return
	}

	if c.pool.idle > 0 && c.lastUsed.Add(c.pool.idle).Before(time.Now()) {
		_ = c.conn.Close()
		return
	}

	if c.pool.ttl > 0 && c.createAt.Add(c.pool.ttl).Before(time.Now()) {
		_ = c.conn.Close()
		return
	}

	c.pool.Lock()
	clients := c.pool.conns[c.addr]
	c.pool.Unlock()

	if clients == nil {
		_ = c.conn.Close()
		return
	}

	select {
	case clients <- c:
	default:
		_ = c.conn.Close()
	}
}
