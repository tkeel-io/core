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

// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation and Dapr Contributors.
// Licensed under the MIT License.
// ------------------------------------------------------------

package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		o := DefaultOptions()
		assert.Equal(t, defaultJSONOutput, o.JSONFormatEnabled)
		assert.Equal(t, defaultAppID, o.appID)
		assert.Equal(t, defaultOutputLevel, o.OutputLevel)
	})

	t.Run("set dapr ID", func(t *testing.T) {
		o := DefaultOptions()
		assert.Equal(t, defaultAppID, o.appID)

		o.SetAppID("dapr-app")
		assert.Equal(t, "dapr-app", o.appID)
	})

	t.Run("attaching log related cmd flags", func(t *testing.T) {
		o := DefaultOptions()

		logLevelAsserted := false
		testStringVarFn := func(p *string, name string, value string, usage string) {
			if name == "log-level" && value == defaultOutputLevel {
				logLevelAsserted = true
			}
		}

		logAsJSONAsserted := false
		testBoolVarFn := func(p *bool, name string, value bool, usage string) {
			if name == "log-as-json" && value == defaultJSONOutput {
				logAsJSONAsserted = true
			}
		}

		o.AttachCmdFlags(testStringVarFn, testBoolVarFn)

		// assert
		assert.True(t, logLevelAsserted)
		assert.True(t, logAsJSONAsserted)
	})
}

func TestApplyOptionsToLoggers(t *testing.T) {
	testOptions := Options{
		JSONFormatEnabled: true,
		appID:             "dapr-app",
		OutputLevel:       "debug",
	}

	// Create two loggers
	testLoggers := []Logger{
		NewLogger("testLogger0"),
		NewLogger("testLogger1"),
	}

	for _, l := range testLoggers {
		l.EnableJSONOutput(false)
		l.SetOutputLevel(InfoLevel)
	}

	assert.NoError(t, ApplyOptionsToLoggers(&testOptions))

	for _, l := range testLoggers {
		ll, _ := l.(*Log)
		assert.Equal(
			t,
			"dapr-app",
			ll.logger.Data[logFieldAppID])
		assert.Equal(
			t,
			toLogrusLevel(DebugLevel),
			ll.logger.Logger.GetLevel())
	}
}
