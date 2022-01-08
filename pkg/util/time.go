package util

import "time"

func UnixMill() int64 {
	return time.Now().UnixNano() / 1e6
}
