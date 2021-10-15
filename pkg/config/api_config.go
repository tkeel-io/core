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
