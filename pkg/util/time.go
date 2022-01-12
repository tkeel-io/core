package util

import "time"

type ElapsedTime struct {
	start time.Time
}

// NewElapsed returns ElapsedTime.
func NewElapsed() ElapsedTime {
	return ElapsedTime{start: time.Now()}
}

// Elapsed returns elapsed duration.
func (et ElapsedTime) Elapsed() time.Duration {
	return time.Since(et.start)
}

// ElapsedSecond returns elapsed seconds.
func (et ElapsedTime) ElapsedSecond() int64 {
	return int64(time.Since(et.start).Seconds())
}

// ElapsedMill returns elapsed milliseconds.
func (et ElapsedTime) ElapsedMilli() int64 {
	return time.Since(et.start).Milliseconds()
}

// ElapsedMicro returns elapsed microseconds.
func (et ElapsedTime) ElapsedMicro() int64 {
	return time.Since(et.start).Microseconds()
}

// ElapsedNano returns elapsed nanoseconds.
func (et ElapsedTime) ElapsedNano() int64 {
	return time.Since(et.start).Nanoseconds()
}

func UnixMilli() int64 {
	return time.Now().UnixNano() / 1e6
}
