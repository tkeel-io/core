package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkeel-io/core/pkg/resource/search/driver"
)

func TestParseSearchEngine(t *testing.T) {
	tests := []struct {
		s    string
		want struct {
			drive    driver.Type
			username string
			password string
			urls     []string
			err      error
		}
	}{
		{"es://username:password@localhost", struct {
			drive    driver.Type
			username string
			password string
			urls     []string
			err      error
		}{drive: driver.DriverTypeElasticsearch, username: "username", password: "password", urls: []string{"localhost"}, err: nil},
		},
		{"es://username:password@localhost,otherhost", struct {
			drive    driver.Type
			username string
			password string
			urls     []string
			err      error
		}{drive: driver.DriverTypeElasticsearch, username: "username", password: "password", urls: []string{"localhost", "otherhost"}, err: nil},
		},
	}

	for _, test := range tests {
		d, u, p, urls, err := ParseSearchEngine(test.s)
		assert.Equal(t, test.want.drive, d)
		assert.Equal(t, test.want.username, u)
		assert.Equal(t, test.want.password, p)
		assert.Equal(t, test.want.urls, urls)
		assert.Equal(t, test.want.err, err)
	}
}
