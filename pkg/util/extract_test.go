package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Extract(t *testing.T) {
	m := map[string]string{
		"field1": "value1",
		"field2": "value2",
		"field3": "value3",
		"field4": "value4",
	}

	assert.Equal(t, "field1=value1,field2=value2,field3=value3,field4=value4", ExtractMap(m))
}
