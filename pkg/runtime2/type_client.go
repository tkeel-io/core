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

package runtime2

import (
	"context"
)

type MessageHandle func(ctx context.Context, message interface{}) error

// Sink
type Sink interface {
	String() string
	Send(ctx context.Context, event interface{}) error
	SendAsync(ctx context.Context, event interface{}) (Promise, error)
	Close(ctx context.Context) error
}

// Source
type Source interface {
	String() string
	StartReceiver(ctx context.Context, fn MessageHandle) error
	Close(ctx context.Context) error
}

type Promise interface {
	Then(s func(err error)) Promise
	Finish(err error) Promise
}
type Receiver interface {
	Receive(context.Context, interface{}) error
}
