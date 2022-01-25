package util

import "testing"

func TestReEntryLock(t *testing.T) {
	requestID := "req-1234"
	lock := NewReEntryLock(12)

	t.Log("entry lock 3 depth.")
	lock.Lock(&requestID)
	lock.Lock(&requestID)
	lock.Lock(&requestID)
	lock.Lock(&requestID)
}
