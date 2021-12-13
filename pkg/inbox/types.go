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

package inbox

const (
	defaultNonBlockNum = 10
	defaultExpiredTime = 300 // ms.

	MsgReciverID             = "m-reciverid"
	MsgReciverStatusActive   = "m-active"
	MsgReciverStatusInactive = "m-inactive"
)

type MessageHandler = func(msg MessageCtx) (int, error)

type Inbox interface {
	Start()
	Stop()
	OnMessage(msg MessageCtx)
}

type Offseter interface {
	Status() bool
	Commit() error
	Confirm()
	AutoCommit() bool
}

type MsgReciver interface {
	Status() string
	OnMessage(msg MessageCtx) (int, error)
}
