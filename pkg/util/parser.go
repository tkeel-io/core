package util

import (
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core/pkg/resource/search/driver"
)

const _splitter = ","

var _searchDriverMap = map[string]driver.Type{
	"es":            driver.ElasticsearchDriver,
	"elasticsearch": driver.ElasticsearchDriver,
}

func ParseSearchEngine(dsn string) (drive driver.Type, username, password string, urls []string, err error) {
	urls = strings.Split(dsn, _splitter)
	u, err := url.Parse(urls[0])
	if err != nil {
		return "", "", "", nil, errors.Wrap(err, "parse search engine dsn error")
	}
	urls[0] = u.Host
	username = u.User.Username()
	password, _ = u.User.Password()
	drive = _searchDriverMap[strings.ToLower(u.Scheme)]
	return drive, username, password, urls, nil
}
