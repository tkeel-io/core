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

package runtime

import (
	"sync"

	"github.com/tkeel-io/core/pkg/statem"
)

type Container struct {
	lock   sync.RWMutex
	states map[string]statem.StateMarchiner
}

func NewContainer() *Container {
	return &Container{
		states: make(map[string]statem.StateMarchiner),
	}
}

func (c *Container) Add(s statem.StateMarchiner) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.states[s.GetID()] = s
}

func (c *Container) Remove(id string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.states, id)
}

func (c *Container) Get(id string) statem.StateMarchiner {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.states[id]
}
