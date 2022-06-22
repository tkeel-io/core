package transport

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
	"unsafe"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

type MockMsg struct {
	value string
}

func (m *MockMsg) Encode() []byte {
	return StrToBytes(m.value)
}

func (m *MockMsg) Decode(b []byte) {
	m.value = ByteToStr(b)
}

func ByteToStr(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func StrToBytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	return *(*[]byte)(unsafe.Pointer(&([3]uintptr{x[0], x[1], x[1]})))
}

func newTimerWrapFunc(t time.Duration, fn func()) {
	tick := time.NewTicker(t)
	ch := make(chan os.Signal, 1)

	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	for counter := 1; counter > 0; counter-- {
		select {
		case <-tick.C:
			fn()
		case <-ch:
			break
		}
	}
	tick.Stop()
	close(ch)
	timer := time.After(time.Second * 7)
	<-timer
}

func TestNewSinkTransport(t *testing.T) {
	sinkTransport, err := NewSinkTransport(
		context.Background(),
		"clickhouse",
		func(msgs []interface{}) (err error) {
			t.Logf("handler msg: %s", msgs)
			return nil
		},
		func(m interface{}) interface{} {
			return m
		})
	if err != nil {
		t.Error(err)
	}
	newTimerWrapFunc(time.Second, func() {
		sinkTransport.Send(context.Background(), &MockMsg{
			value: "1111",
		})
	})
}
