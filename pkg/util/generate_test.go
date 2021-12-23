package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FormatMapper(t *testing.T) {
	assert.Equal(t, "core.BASIC.mapper.device123.mapper123", FormatMapper("BASIC", "device123", "mapper123"))
}
