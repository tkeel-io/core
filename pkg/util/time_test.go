package util

import (
	"testing"
	"time"
)

func Test_NewElapsed(t *testing.T) {
	elapsedTime := NewElapsed()

	time.Sleep(1 * time.Second)
	t.Log("elapsed seconds: ", elapsedTime.ElapsedSecond())
	t.Log("elapsed milliseconds: ", elapsedTime.ElapsedMilli())
	t.Log("elapsed microseconds: ", elapsedTime.ElapsedMicro())
	t.Log("elapsed nanoseconds: ", elapsedTime.ElapsedNano())
}

func Test_ElapsedSecond(t *testing.T) {
	elapsedTime := NewElapsed()

	time.Sleep(1 * time.Second)
	t.Log("elapsed seconds: ", elapsedTime.ElapsedSecond())
}

func Test_ElapsedMilli(t *testing.T) {
	elapsedTime := NewElapsed()

	time.Sleep(3 * time.Millisecond)
	t.Log("elapsed milliseconds: ", elapsedTime.ElapsedMilli())
}

func Test_ElapsedMicro(t *testing.T) {
	elapsedTime := NewElapsed()

	time.Sleep(3 * time.Microsecond)
	t.Log("elapsed microseconds: ", elapsedTime.ElapsedMicro())
}

func Test_ElapsedNano(t *testing.T) {
	elapsedTime := NewElapsed()

	time.Sleep(3 * time.Nanosecond)
	t.Log("elapsed nanoseconds: ", elapsedTime.ElapsedNano())
}
