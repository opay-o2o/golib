package gorm

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/opay-o2o/golib/logger"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Host         string `toml:"host"`
	Port         uint   `toml:"port"`
	User         string `toml:"user"`
	Password     string `toml:"password"`
	Charset      string `toml:"charset"`
	Database     string `toml:"database"`
	Timeout      int    `toml:"timeout" json:"timeout"`
	MaxOpenConns int    `toml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns int    `toml:"max_idle_conns" json:"max_idle_conns"`
	MaxConnTtl   int    `toml:"max_conn_ttl" json:"max_conn_ttl"`
	Debug        bool   `toml:"debug"`
}

func (c *Config) GetDsn() string {
	if c.Timeout <= 0 {
		c.Timeout = 3
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&timeout=%ds",
		c.User, c.Password, c.Host, c.Port, c.Database, c.Charset, c.Timeout)
}

type Logger struct {
	logger *logger.Logger
}

func (l *Logger) Print(values ...interface{}) {
	if len(values) > 1 {
		source := values[1].(string)

		if dirs := strings.Split(source, "/"); len(dirs) >= 3 {
			source = strings.Join(dirs[len(dirs)-3:], "/")
		}

		if values[0] == "sql" {
			if len(values) > 5 {
				sql := gorm.LogFormatter(values...)[3]
				execTime := float64(values[2].(time.Duration).Nanoseconds()/1e4) / 100.0
				rows := values[5].(int64)
				l.logger.Debugf("query: <%s> | %.2fms | %d rows | %s", source, execTime, rows, sql)
			}
		} else {
			l.logger.Debug(source, values[2:])
		}
	}
}

type Pool struct {
	locker  sync.RWMutex
	clients map[string]*gorm.DB
	logger  *logger.Logger
}

func (p *Pool) Add(name string, c *Config) error {
	p.locker.Lock()
	defer p.locker.Unlock()

	orm, err := gorm.Open("mysql", c.GetDsn())

	if err != nil {
		return err
	}

	db := orm.DB()

	if c.MaxIdleConns > 0 {
		db.SetMaxIdleConns(c.MaxIdleConns)
	}

	if c.MaxOpenConns > 0 {
		db.SetMaxOpenConns(c.MaxOpenConns)
	}

	if c.MaxConnTtl > 0 {
		db.SetConnMaxLifetime(time.Duration(c.MaxConnTtl) * time.Second)
	}

	if c.Debug {
		orm.LogMode(true)
	}

	orm.SetLogger(&Logger{p.logger})
	p.clients[name] = orm

	return nil
}

func (p *Pool) Get(name string) (*gorm.DB, error) {
	p.locker.RLock()
	defer p.locker.RUnlock()

	client, ok := p.clients[name]

	if ok {
		return client, nil
	}

	return nil, errors.New("no mysql gorm client")
}

func NewPool(logger *logger.Logger) *Pool {
	return &Pool{clients: make(map[string]*gorm.DB, 64), logger: logger}
}
