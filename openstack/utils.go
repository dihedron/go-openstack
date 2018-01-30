package openstack

import (
	"strings"
)

// NormaliseURL ensures that the returned URL always ends with a "/"; this
// is necessary because service base URLs in Slings must have a trailing "/"
// for relative paths to be correctly constructed.
func NormaliseURL(url string) string {
	if strings.HasSuffix(url, "/") {
		return url
	}
	return url + "/"
}
