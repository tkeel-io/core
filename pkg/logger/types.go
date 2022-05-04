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

package logf

import "go.uber.org/zap"

var (
	Any           = zap.Any
	Error         = zap.Error
	String        = zap.String
	Float32       = zap.Float32
	Float64       = zap.Float64
	Float32s      = zap.Float32s
	Float64s      = zap.Float64s
	Int           = zap.Int
	Int32         = zap.Int32
	Int64         = zap.Int64
	Bool          = zap.Bool
	ByteString    = zap.ByteString
	Time          = zap.Time
	Fields        = zap.Fields
	AddStacktrace = zap.AddStacktrace
)
