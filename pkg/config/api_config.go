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

package config

type APIConfig struct {
	EventAPIConfig      EventAPIConfig      `mapstructure:"event_api_config"`
	EntityAPIConfig     EntityAPIConfig     `mapstructure:"entity_api_config"`
	TimeSeriesAPIConfig TimeSeriesAPIConfig `mapstructure:"time_series_api_config"`
}

type EventAPIConfig struct {
	RawTopic          string `mapstructure:"raw_topic"`
	TimeSeriesTopic   string `mapstructure:"time_series_topic"`
	PropertyTopic     string `mapstructure:"property_topic"`
	RelationshipTopic string `mapstructure:"relationship_topic"`
	StoreName         string `mapstructure:"store_name"`
	PubsubName        string `mapstructure:"pubsub_name"`
}

type TimeSeriesAPIConfig struct {
}

type EntityAPIConfig struct {
	TableName   string `mapstructure:"table_name"`
	StateName   string `mapstructure:"state_name"`
	BindingName string `mapstructure:"binding_name"`
}
