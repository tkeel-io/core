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
	_defaultProxy = Proxy{
		Name:     "core.proxy",
		HTTPPort: 7000,
		GRPCPort: 7001,
	}
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

func SetDefaultEtcd(etcdBrokers []string) {
	if len(etcdBrokers) > 0 {
		_defaultEtcdConfig.Endpoints = etcdBrokers
		_defaultDiscovery.Endpoints = etcdBrokers
	}
}

func SetDefaultES(esBrokers []string) {
	_defaultESConfig.Endpoints = esBrokers
}
