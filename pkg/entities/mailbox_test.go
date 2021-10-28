package entities

import (
	"math/rand"
	"sync"
	"testing"
)

func TestMailBox(t *testing.T) {
	mb := newMailbox(5)
	wg := sync.WaitGroup{}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(num int) {
			op := "PUT"
			if num%2 == 0 {
				for j := 0; j < num; j++ {
					mb.Put(nil)
				}
			} else {
				op = "GET"
				for j := 0; j < num; j++ {
					mb.Get()
				}
			}

			t.Logf("%s messages %d.", op, num)
			wg.Done()
		}(rand.Intn(100 * 10000)) //nolint
	}

	wg.Wait()
}
