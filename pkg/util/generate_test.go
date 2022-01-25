package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FormatMapper(t *testing.T) {
	// format print.
	assert.Equal(t, "core.mapper.BASIC.device123.mapper123", FormatMapper("BASIC", "device123", "mapper123"))
}
