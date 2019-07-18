package grpc

import (
	"context"
	"fmt"
	"github.com/opay-o2o/golib/net2"
	"google.golang.org/grpc/metadata"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var serialNumber uint64
var localIp string

func GenTraceId() string {
	ip := "0.0.0.0"

	if localIp == "" {
		if lip, err := net2.GetLocalIp(); err == nil {
			ip = lip
		}
	}

	ips := strings.Split(ip, ".")
	params := []interface{}{0, 0, 0, 0}

	for k, v := range ips {
		if k < len(params) {
			if n, err := strconv.Atoi(v); err == nil {
				params[k] = n
			}
		}
	}

	params = append(params, time.Now().UnixNano()/1e6, atomic.AddUint64(&serialNumber, 1))
	return fmt.Sprintf("%02x%02x%02x%02x%013d%04d", params...)
}

func NewTraceCtx(parent context.Context) context.Context {
	var trackId string

	if parent != nil {
		trackId = GetCtxTraceId(parent)
	}

	if trackId == "" {
		trackId = GenTraceId()
	}

	return metadata.NewOutgoingContext(
		context.Background(),
		metadata.New(map[string]string{"trace_id": trackId}),
	)
}

func GetCtxTraceId(ctx context.Context) (trackId string) {
	data, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return
	}

	v, ok := data["trace_id"]

	if !ok || len(v) < 1 {
		return
	}

	if v[0] != "" {
		trackId = v[0]
	}

	return
}
