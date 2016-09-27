package api

import "regexp"

var idRegexp = regexp.MustCompile("^[0-9a-z](?:[_-]?[0-9a-z]){1,35}$")

// ValidID returns true if the given ID is a valid application or device ID
func ValidID(id string) bool {
	return idRegexp.Match([]byte(id))
}
