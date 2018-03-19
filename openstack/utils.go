package openstack

import (
	"fmt"
	"strings"
)

// NormaliseURL ensures that the returned URL always ends with a "/"; this
// is necessary because service base URLs in Builders must have a trailing "/"
// for relative paths to be correctly constructed.
func NormaliseURL(url string) string {
	if strings.HasSuffix(url, "/") {
		return url
	}
	return url + "/"
}

// ZipString returns a truncated version of the input string, with the given
// length, where the middle characters are replaced with three dots ("...").
func ZipString(s string, length int) string {
	switch {
	case length <= 2:
		return ""
	case length == 3:
		return "..."
	case length%2 == 1:
		half := (length - 3) / 2
		return fmt.Sprintf("%s...%s", s[0:half], s[len(s)-half:])
	case length%2 == 0:
		head := ((length - 3) + 1) / 2
		tail := ((length - 3) - 1) / 2
		return fmt.Sprintf("%s...%s", s[0:head], s[len(s)-tail:])
	}
	return ""
}

// func ISO8601ToTime(date string) (time.Time, error) {
// 	return time.Parse(ISO8601, date)
// }

// func TimeToISO8601(time time.Time) (string, error) {
// 	return time.Format(ISO8601)
// }

// for i := -1; i < 10; i++ {
// 	fmt.Printf("%d: %s\n", i, openstack.ZipString("abcdefghijklmnopqrstuvwxyz", i))
// }
//fmt.Println("%s", openstack.ZipString("abcdefghijklmnopqrstuvwxyz", 5))ùreturn
