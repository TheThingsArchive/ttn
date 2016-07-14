package api

import "regexp"

var idRegexp = regexp.MustCompile("^[[:alnum:]](?:[_-]?[[:alnum:]]){1,35}$")

// ValidID returns true if the given ID is a valid application or device ID
func ValidID(id string) bool {
	return idRegexp.Match([]byte(id))
}
