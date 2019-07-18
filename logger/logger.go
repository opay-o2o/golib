package logger

import (
	"../net2"
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Dir        string `toml:"dir"`
	Level      string `toml:"level"`
	Color      bool   `toml:"color"`
	Terminal   bool   `toml:"terminal"`
	ShowIp     bool   `toml:"show_ip"`
	TimeFormat string `toml:"time_format"`
}

func DefaultConfig() *Config {
	return &Config{
		Dir:        "./logs",
		Level:      "debug",
		Color:      true,
		Terminal:   true,
		ShowIp:     true,
		TimeFormat: "2006-01-02 15:04:05",
	}
}

type Logger struct {
	c        *Config
	level    Level
	file     string
	ip       string
	f        *os.File
	w        *bufio.Writer
	bytePool *sync.Pool
	ch       chan interface{}
	timer    *time.Ticker
	end      chan bool
}

type msg struct {
	file   string
	line   int
	level  Level
	format string
	args   []interface{}
}

func NewLogger(c *Config) (*Logger, error) {
	l := &Logger{
		c:        c,
		level:    GetLevel(c.Level),
		bytePool: &sync.Pool{New: func() interface{} { return new(bytes.Buffer) }},
		ch:       make(chan interface{}, 8192),
		timer:    time.NewTicker(time.Second),
		end:      make(chan bool, 1),
	}

	err := os.MkdirAll(l.c.Dir, 0755)

	if err != nil {
		return nil, err
	}

	l.ip, err = net2.GetLocalIp()

	if err != nil {
		return nil, err
	}

	l.refresh()
	l.f, err = os.OpenFile(l.file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return nil, err
	}

	l.w = bufio.NewWriter(l.f)
	l.run()

	return l, nil
}

func (l *Logger) refresh() bool {
	oldFile := l.file
	l.file = path.Join(l.c.Dir, time.Now().Format("20060102.log"))
	return l.file != oldFile
}

func (l *Logger) start() {
	for m := range l.ch {
		if m == nil {
			l.w.Flush()

			if l.refresh() {
				l.f.Close()
				l.f, _ = os.OpenFile(l.file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				l.w.Reset(l.f)
			}
		} else if msg, ok := m.(*msg); ok {
			l.w.Write(l.bytes(msg))
		} else if p, ok := m.([]byte); ok {
			l.w.Write(p)
		}
	}

	l.end <- true
}

func (l *Logger) run() {
	go l.flush()
	go l.start()
}

func (l *Logger) bytes(m *msg) []byte {
	w := l.bytePool.Get().(*bytes.Buffer)

	defer func() {
		recover()
		w.Reset()
		l.bytePool.Put(w)
	}()

	nowTime := time.Now().Format(l.c.TimeFormat)
	level := GetLevelText(m.level, l.c.Color)
	loc := fmt.Sprintf("<%s:%d>", m.file, m.line)

	if l.c.Color {
		loc = Blue(loc)
	}

	if l.c.ShowIp {
		fmt.Fprintf(w, "%s (%s) %s %s ", level, l.ip, nowTime, loc)
	} else {
		fmt.Fprintf(w, "%s %s %s ", level, nowTime, loc)
	}

	if len(m.format) == 0 {
		for i := 0; i < len(m.args); i++ {
			if i > 0 {
				w.Write([]byte{' '})
			}

			fmt.Fprint(w, m.args[i])
		}
	} else {
		fmt.Fprintf(w, m.format, m.args...)
	}

	fmt.Fprintf(w, "\n")

	b := make([]byte, w.Len())
	copy(b, w.Bytes())

	return b
}

func (l *Logger) flush() {
	for range l.timer.C {
		l.ch <- nil
	}
}

func (l *Logger) getFileInfo() (file string, line int) {
	_, file, line, ok := runtime.Caller(3)

	if !ok {
		return "???", 1
	}

	if dirs := strings.Split(file, "/"); len(dirs) >= 2 {
		return dirs[len(dirs)-2] + "/" + dirs[len(dirs)-1], line
	}

	return
}

func (l *Logger) Log(level Level, format string, args ...interface{}) {
	m := &msg{level: level, format: format, args: args}
	m.file, m.line = l.getFileInfo()

	if l.c.Terminal {
		fmt.Fprint(os.Stdout, string(l.bytes(m)))
	} else {
		if level <= l.level {
			l.ch <- m
		}
	}
}

func (l *Logger) Write(p []byte) (n int, err error) {
	l.ch <- p
	return len(p), nil
}

func (l *Logger) Debug(args ...interface{}) {
	l.Log(DebugLevel, "", args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Log(DebugLevel, format, args...)
}

func (l *Logger) Info(args ...interface{}) {
	l.Log(InfoLevel, "", args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.Log(InfoLevel, format, args...)
}

func (l *Logger) Warning(args ...interface{}) {
	l.Log(WarnLevel, "", args...)
}

func (l *Logger) Warningf(format string, args ...interface{}) {
	l.Log(WarnLevel, format, args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.Log(ErrorLevel, "", args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Log(ErrorLevel, format, args...)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.Log(FatalLevel, "", args...)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Log(FatalLevel, format, args...)
}

func (l *Logger) Close() {
	l.timer.Stop()
	l.ch <- nil
	close(l.ch)
	<-l.end
}
