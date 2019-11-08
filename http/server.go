package http

import (
	stdContext "context"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/middleware/pprof"
	"github.com/kataras/iris/v12/websocket"
	"github.com/kataras/neffos"
	"github.com/opay-o2o/golib/logger"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type WsConfig struct {
	Enable   bool          `toml:"enable"`
	Endpoint string        `toml:"endpoint"`
	IdleTime time.Duration `toml:"idle_time"`
}

type LogConfig struct {
	Level      string `toml:"level"`
	TimeFormat string `toml:"time_format"`
	Color      bool   `toml:"color"`
}

type TlsConfig struct {
	Enable   bool   `toml:"enable"`
	CertPath string `toml:"cert_path"`
	KeyPath  string `toml:"key_path"`
}

type Config struct {
	Host      string     `toml:"host"`
	Port      int        `toml:"port"`
	Charset   string     `toml:"charset"`
	Gzip      bool       `toml:"gzip"`
	PProf     bool       `toml:"pprof"`
	Websocket *WsConfig  `toml:"websocket"`
	Tls       *TlsConfig `toml:"tls"`
	Log       *LogConfig `toml:"log"`
}

func (c *Config) GetAddr() string {
	return c.Host + ":" + strconv.Itoa(c.Port)
}

func DefaultConfig() *Config {
	return &Config{
		Port:    80,
		Charset: "UTF-8",
		Websocket: &WsConfig{
			Enable: false,
		},
		Tls: &TlsConfig{
			Enable: false,
		},
		Log: &LogConfig{
			Level:      "debug",
			TimeFormat: "2006-01-02 15:04:05",
			Color:      true,
		},
	}
}

func GetClientIp(ctx context.Context) string {
	xForwarded := ctx.GetHeader("X-Forwarded-For")

	if ip := strings.TrimSpace(strings.Split(xForwarded, ",")[0]); ip != "" {
		return ip
	}

	if xReal := strings.TrimSpace(ctx.GetHeader("X-Real-Ip")); xReal != "" {
		return xReal
	}

	return ctx.RemoteAddr()
}

type Router interface {
	RegHttpHandlers(app *iris.Application)
	WebsocketConnectHandler(conn *neffos.Conn) error
	WebsocketDisconnectHandler(conn *neffos.Conn)
	WebsocketMessageHandler(conn *websocket.NSConn, msg websocket.Message) error
	GetIdentifier(ctx context.Context) string
}

type Server struct {
	sync.Mutex
	config   *Config
	router   Router
	app      *iris.Application
	ws       *neffos.Server
	logger   *logger.Logger
	ctx      stdContext.Context
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

// recovery panic (500)
func (s *Server) Recovery(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			if ctx.IsStopped() {
				return
			}

			var stacktrace string

			for i := 1; ; i++ {
				_, f, l, got := runtime.Caller(i)

				if !got {
					break
				}

				stacktrace += fmt.Sprintf("%s:%d\n", f, l)
			}

			request := fmt.Sprintf("%v %s %s %s", strconv.Itoa(ctx.GetStatusCode()), GetClientIp(ctx), ctx.Method(), ctx.Path())
			s.logger.Error(fmt.Sprintf("recovered panic:\nRequest: %s\nTrace: %s\n%s", request, err, stacktrace))

			ctx.StatusCode(500)
			ctx.StopExecution()
		}
	}()

	ctx.Next()
}

// record access log
func (s *Server) AccessLog(ctx context.Context) {
	start := time.Now()
	ctx.Next()

	idf := s.router.GetIdentifier(ctx)
	statusCode, useTime, clientIp := ctx.GetStatusCode(), time.Since(start), GetClientIp(ctx)
	uri, method, userAgent := ctx.Request().URL.RequestURI(), ctx.Method(), ctx.GetHeader("User-Agent")
	s.logger.Infof("request: %d | %4v | %s | %s %s | %s | %s", statusCode, useTime, clientIp, method, uri, userAgent, idf)
}

func CrossDomain(ctx context.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Next()
}

func UnGzip(ctx context.Context) {
	ctx.Gzip(false)
	ctx.Next()
}

func (s *Server) Start() {
	go func() {
		var runner iris.Runner

		if s.config.Tls.Enable {
			runner = iris.TLS(s.config.GetAddr(), s.config.Tls.CertPath, s.config.Tls.KeyPath)
		} else {
			runner = iris.Addr(s.config.GetAddr())
		}

		err := s.app.Run(runner, iris.WithConfiguration(iris.Configuration{
			DisableStartupLog:                 true,
			DisableInterruptHandler:           true,
			DisableBodyConsumptionOnUnmarshal: true,
			Charset:                           s.config.Charset,
		}))

		if err != nil && s.Running() {
			s.logger.Errorf("can't serve at <%s> | error: %s", s.config.GetAddr(), err)
		}
	}()
}

func (s *Server) Stop() {
	s.canceler()

	ctx, _ := stdContext.WithTimeout(stdContext.Background(), 3*time.Second)

	if err := s.app.Shutdown(ctx); err != nil {
		s.logger.Errorf("server shutdown error: %s", err)
	}
}

func (s *Server) GetWsConn(connId string) *neffos.Conn {
	return s.ws.GetConnections()[connId]
}

func NewServer(c *Config, r Router, l *logger.Logger) *Server {
	server := &Server{config: c, router: r, logger: l}
	server.ctx, server.canceler = stdContext.WithCancel(stdContext.Background())

	// create iris instance
	server.app = iris.New()
	server.app.Use(server.Recovery)
	server.app.Use(server.AccessLog)

	// enable gzip
	if c.Gzip {
		server.app.Use(iris.Gzip)
	}

	// enable pprof
	if c.PProf {
		server.app.Any("/debug/pprof", pprof.New())
		server.app.Any("/debug/pprof/{action:path}", pprof.New())
	}

	// set logger
	server.app.Logger().SetLevel(c.Log.Level)
	server.app.Logger().SetTimeFormat(c.Log.TimeFormat)
	server.app.Logger().SetOutput(l)
	server.app.Logger().Printer.IsTerminal = c.Log.Color

	// set route
	server.router.RegHttpHandlers(server.app)

	// set websocket
	if c.Websocket.Enable {
		server.ws = websocket.New(websocket.DefaultGorillaUpgrader, neffos.WithTimeout{
			ReadTimeout: c.Websocket.IdleTime * time.Second,
			Events: neffos.Events{
				websocket.OnNativeMessage: server.router.WebsocketMessageHandler,
			},
		})

		server.ws.OnConnect = server.router.WebsocketConnectHandler
		server.ws.OnDisconnect = server.router.WebsocketDisconnectHandler
		server.app.Get(c.Websocket.Endpoint, websocket.Handler(server.ws))
	}

	return server
}
