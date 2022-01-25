package util

import (
	"sort"
	"strings"
)

func ExtractMap(m map[string]string) string {
	var pairs sort.StringSlice
	for key, value := range m {
		pairs = append(pairs, key+"="+value)
	}

	sort.Sort(pairs)
	return strings.Join([]string(pairs), ",")
}
