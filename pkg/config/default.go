package config

const (
	_defaultConfigFilename = "config.yml"
	_corePrefix            = "CORE"

	// default app id.
	DefaultAppID = "core"
	// assume single node.
	DefaultName = "core"
)

var (
	_defaultProxyConfig = Proxy{
		HTTPPort: 20000,
		GRPCPort: 20001,
	}
	_defaultAppServer = Server{
		Name:     DefaultName,
		AppID:    DefaultAppID,
		HTTPAddr: ":6789",
		GRPCAddr: ":31234",
	}
	_defaultLogConfig = LogConfig{
		Dev:      true,
		Level:    "INFO",
		Encoding: "json",
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
