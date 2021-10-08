package config

type ApiConfig struct {
	EventApiConfig      EventApiConfig      `mapstructure:"event_api_config"`
	EntityApiConfig     EntityApiConfig     `mapstructure:"entity_api_config"`
	TimeSeriesApiConfig TimeSeriesApiConfig `mapstructure:"time_series_api_config"`
}

type EventApiConfig struct {
	RawTopic          string `mapstructure:"raw_topic"`
	TimeSeriesTopic   string `mapstructure:"time_series_topic"`
	PropertyTopic     string `mapstructure:"property_topic"`
	RelationShipTopic string `mapstructure:"relationship_topic"`
	StoreName         string `mapstructure:"store_name"`
	PubsubName        string `mapstructure:"pubsub_name"`
}

type TimeSeriesApiConfig struct {
}

type EntityApiConfig struct {
	TableName   string `mapstructure:"table_name"`
	StateName   string `mapstructure:"state_name"`
	BindingName string `mapstructure:"binding_name"`
}
