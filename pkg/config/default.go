package config

const (
	_httpScheme            = "http"
	_schemeSpliterator     = "://"
	_defaultConfigFilename = "config.yml"
	_corePrefix            = "CORE"

	// default app port.
	DefaultAppPort = 6789
	// default app id.
	DefaultAppID = "core"
	// assume single node.
	DefaultName = "core"
)

var (
	_defaultAppServer = Server{
		Name:    DefaultName,
		AppID:   DefaultAppID,
		AppPort: DefaultAppPort,
	}
	_defaultLogConfig = LogConfig{
		Level:    "INFO",
		Encoding: "json",
	}
	_defaultUseSearchEngine = "elasticsearch"
	_defaultESConfig        = ESConfig{
		Endpoints: []string{"http://localhost:9200"},
		Username:  "admin",
		Password:  "admin",
	}
	_defaultEtcdConfig = EtcdConfig{
		DialTimeout: 3,
		Endpoints:   []string{"http://localhost:2379"},
	}
	_defaultDiscovery = Discovery{
		HeartTime:   3,
		DialTimeout: 3,
		Endpoints:   []string{"http://localhost:2379"},
	}
)
