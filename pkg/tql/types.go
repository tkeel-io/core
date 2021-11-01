package tql

/*
   1. 对MQL的静态分析
   	1. 解析Target Entity
   	2. 解析Source Entities
   	3. 解析Tentacles
   2. MQL运行时执行
   	1. json输入
   	2. 执行输出json
   	3. 输出的json反馈给Target
*/

type TentacleConfig struct {
	SourceEntity string
	PropertyKeys []string
}

type TQLConfig struct { // nolint
	TargetEntity   string
	SourceEntities []string
	Tentacles      []TentacleConfig
}

type TQL interface {
	Target() string
	Entities() []string
	Tentacles() []TentacleConfig
	Exec(map[string]interface{}) (map[string]interface{}, error)
	ExecJSONE([]byte) (map[string]interface{}, error)
}
