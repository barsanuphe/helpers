package helpers

import "strings"

// StringInSlice checks if a string is in a []string.
func StringInSlice(a string, list []string) (index int, isIn bool) {
	for i, b := range list {
		if b == a {
			return i, true
		}
	}
	return -1, false
}

// RemoveDuplicates in []string
func RemoveDuplicates(options *[]string, otherStringsToClean ...string) {
	found := make(map[string]bool)
	// specifically remove other strings from values
	for _, o := range otherStringsToClean {
		found[o] = true
	}
	j := 0
	for i, x := range *options {
		if !found[x] && x != "" {
			found[x] = true
			(*options)[j] = (*options)[i]
			j++
		}
	}
	*options = (*options)[:j]
}

// StringInSliceCaseInsensitive checks if a string is in a []string, regardless of case.
func StringInSliceCaseInsensitive(a string, list []string) (index int, isIn bool) {
	for i, b := range list {
		if strings.ToLower(b) == strings.ToLower(a) {
			return i, true
		}
	}
	return -1, false
}

// CaseInsensitiveContains checks if a substring is in a string, regardless of case.
func CaseInsensitiveContains(s, substr string) bool {
	s, substr = strings.ToLower(s), strings.ToLower(substr)
	return strings.Contains(s, substr)
}
