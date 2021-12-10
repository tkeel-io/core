/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package statem

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
