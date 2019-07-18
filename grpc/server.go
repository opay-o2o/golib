package grpc

import (
	"context"
	"github.com/opay-o2o/golib/logger"
	"google.golang.org/grpc"
	"net"
	"strconv"
)

type Config struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`
}

func (c *Config) GetAddr() string {
	return c.Host + ":" + strconv.Itoa(c.Port)
}

type Router interface {
	RegGrpcService(server *grpc.Server)
}

type Server struct {
	config   *Config
	server   *grpc.Server
	router   Router
	logger   *logger.Logger
	ctx      context.Context
	canceler func()
}

func (s *Server) Running() bool {
	select {
	case <-s.ctx.Done():
		return false
	default:
		return true
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.config.GetAddr())

	if err != nil {
		return err
	}

	s.server = grpc.NewServer()
	s.router.RegGrpcService(s.server)
	s.ctx, s.canceler = context.WithCancel(context.Background())

	go func() {
		err = s.server.Serve(listener)

		if err != nil && s.Running() {
			s.logger.Errorf("can't serve at <%s>", s.config.GetAddr())
		}
	}()

	return nil
}

func (s *Server) Stop() {
	s.canceler()
	s.server.Stop()
}

func NewServer(c *Config, r Router, l *logger.Logger) *Server {
	return &Server{config: c, router: r, logger: l}
}
