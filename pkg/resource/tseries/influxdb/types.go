package influxdb

import "errors"

var (
	ErrInfluxRequiredURL    = errors.New("Influx Error: URL required")
	ErrInfluxRequiredToken  = errors.New("Influx Error: Token required")
	ErrInfluxRequiredOrg    = errors.New("Influx Error: Org required")
	ErrInfluxRequiredBucket = errors.New("Influx Error: Bucket required")
	ErrInfluxInvalidParams  = errors.New("Influx Error: Cannot convert request data")
)
