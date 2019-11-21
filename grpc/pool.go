package grpc

import (
	"context"
	"google.golang.org/grpc"
	"time"
)

type Dialer func(addr string) (*grpc.ClientConn, error)

func DefaultDialer(addr string) (*grpc.ClientConn, error) {
	return grpc.Dial(addr, grpc.WithInsecure())
}

type Pool struct {
	addr    string
	clients chan *Connection
	dialer  Dialer
	idle    time.Duration
	ttl     time.Duration
}

type Connection struct {
	conn     *grpc.ClientConn
	pool     *Pool
	createAt time.Time
	lastUsed time.Time
}

func NewPool(dialer Dialer, addr string, init, capacity int, idle time.Duration, ttl ...time.Duration) (*Pool, error) {
	if capacity <= 0 {
		capacity = 1
	}

	if init < 0 {
		init = 0
	}

	if init > capacity {
		init = capacity
	}

	p := &Pool{
		addr:    addr,
		clients: make(chan *Connection, capacity),
		dialer:  dialer,
		idle:    idle,
	}

	if len(ttl) > 0 {
		p.ttl = ttl[0]
	}

	for i := 0; i < init; i++ {
		c, err := dialer(addr)

		if err != nil {
			return nil, err
		}

		p.clients <- &Connection{
			pool:     p,
			conn:     c,
			createAt: time.Now(),
			lastUsed: time.Now(),
		}
	}

	return p, nil
}

func (p *Pool) Get(ctx context.Context) (*Connection, error) {
	for {
		select {
		case client := <-p.clients:
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
			c, err := p.dialer(p.addr)

			if err != nil {
				return nil, err
			}

			client := &Connection{
				pool:     p,
				conn:     c,
				createAt: time.Now(),
				lastUsed: time.Now(),
			}

			return client, nil
		}
	}
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

	select {
	case c.pool.clients <- c:
	default:
	}
}
