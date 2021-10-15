package mql

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

type MQL interface {
	Target() string
	Entities() []string
	Tentacles() map[string][]string
	Exec(map[string]map[string]interface{}) (map[string]map[string]interface{}, error)
	ExecJSONE([]byte) (map[string]map[string]interface{}, error)
}
