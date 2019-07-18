package logger

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path"
	"sync"
	"time"
)

const (
	PartitionDay = iota
	PartitionHour
)

type DataConfig struct {
	Dir       string         `toml:"dir"`
	Partition int            `toml:"partition"`
	Timezone  string         `toml:"timezone"`
	Location  *time.Location `toml:"-"`
}

type DataLogger struct {
	c        *DataConfig
	file     string
	f        *os.File
	w        *bufio.Writer
	bytePool *sync.Pool
	ch       chan *data
	timer    *time.Ticker
	end      chan bool
}

type data struct {
	format string
	args   []interface{}
}

func NewDataLogger(c *DataConfig) (*DataLogger, error) {
	l := &DataLogger{
		c:        c,
		bytePool: &sync.Pool{New: func() interface{} { return new(bytes.Buffer) }},
		ch:       make(chan *data, 8192),
		timer:    time.NewTicker(time.Second),
		end:      make(chan bool, 1),
	}

	err := os.MkdirAll(l.c.Dir, 0755)

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

func (l *DataLogger) refresh() bool {
	oldFile := l.file
	nowTime := time.Now().In(l.c.Location)
	year, month, day := nowTime.Date()
	hour := nowTime.Hour()

	switch l.c.Partition {
	case PartitionHour:
		time.Now().Year()
		l.file = path.Join(l.c.Dir, fmt.Sprintf("%04d%02d%02d.%02d.log", year, month, day, hour))
	default:
		l.file = path.Join(l.c.Dir, fmt.Sprintf("%04d%02d%02d.log", year, month, day))
	}

	return l.file != oldFile
}

func (l *DataLogger) start() {
	for m := range l.ch {
		if m == nil {
			l.w.Flush()

			if l.refresh() {
				l.f.Close()
				l.f, _ = os.OpenFile(l.file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				l.w.Reset(l.f)
			}
		} else {
			l.w.Write(l.bytes(m))
		}
	}

	l.end <- true
}

func (l *DataLogger) run() {
	go l.flush()
	go l.start()
}

func (l *DataLogger) bytes(m *data) []byte {
	w := l.bytePool.Get().(*bytes.Buffer)

	defer func() {
		recover()
		w.Reset()
		l.bytePool.Put(w)
	}()

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

func (l *DataLogger) flush() {
	for range l.timer.C {
		l.ch <- nil
	}
}

func (l *DataLogger) log(format string, args ...interface{}) {
	m := &data{format: format, args: args}

	if l.f == nil {
		fmt.Fprint(os.Stdout, l.bytes(m))
	} else {
		l.ch <- m
	}
}

func (l *DataLogger) Log(args ...interface{}) {
	l.log("", args...)
}

func (l *DataLogger) Logf(format string, args ...interface{}) {
	l.log(format, args...)
}

func (l *DataLogger) Close() {
	l.timer.Stop()
	l.ch <- nil
	close(l.ch)
	<-l.end
}
