package path

import (
	"errors"
	"regexp"
	"strings"
)

// ErrZeroLength is returned by Parse if a paths has a zero length.
var ErrZeroLength = errors.New("zero length path")

// ErrWildcards is returned by Parse if a path contains invalid wildcards.
var ErrWildcards = errors.New("invalid use of wildcards")

var multiSlashRegex = regexp.MustCompile(`/+`)

// Parse removes duplicate and trailing slashes from the supplied.
// string and returns the normalized path.
func Parse(path string, allowWildcards bool) (string, error) {
	// check for zero length.
	if path == "" {
		return "", ErrZeroLength
	}

	// normalize path.
	path = multiSlashRegex.ReplaceAllString(path, "/")

	// remove trailing slashes.
	path = strings.TrimRight(path, "/")

	// check again for zero length.
	if path == "" {
		return "", ErrZeroLength
	}

	// split to segments.
	segments := strings.Split(path, "/")

	// check all segments.
	for i, s := range segments {
		// check use of wildcards.
		if (strings.Contains(s, "+") || strings.Contains(s, "#")) && len(s) > 1 {
			return "", ErrWildcards
		}

		// check if wildcards are allowed.
		if !allowWildcards && (s == "#" || s == "+") {
			return "", ErrWildcards
		}

		// check if hash is the last character.
		if s == "#" && i != len(segments)-1 {
			return "", ErrWildcards
		}
	}

	return path, nil
}

// ContainsWildcards tests if the supplied path contains wildcards. The paths.
// is expected to be tested and normalized using Parse beforehand.
func ContainsWildcards(path string) bool {
	return strings.Contains(path, "+") || strings.Contains(path, "#")
}
