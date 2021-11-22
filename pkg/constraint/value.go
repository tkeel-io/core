package constraint

type Value struct {
	Value    []byte
	Config   Config
	LastTime int64
}

type Config struct {
	item

	ID                string `json:"id"`
	Type              string `json:"type"`      // 用于描述entity运行时的属性值的结构信息.
	DataType          string `json:"data_type"` // 用于描述entity运行时属性值的存在形式，默认[]byte.
	Weight            int    `json:"weight"`
	Enabled           bool   `json:"enabled"`
	EnabledSearch     bool   `json:"enabled_search"`
	EnabledTimeSeries bool   `json:"enabled_time_series"`
	Description       string `json:"description"`
	Define            Define `json:"define"`
	LastTime          int64  `json:"last_time"`
}

type SetConfigRequest struct {
	ID                string `json:"id"`
	Type              string `json:"type"`
	Weight            int    `json:"weight"`
	Enabled           bool   `json:"enabled"`
	EnabledSearch     bool   `json:"enabled_search"`
	EnabledTimeSeries bool   `json:"enabled_time_series"`
	Description       string `json:"description"`
	Define            Define `json:"define"`
}
