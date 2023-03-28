package utils

import (
	"strings"
)

// CheckStringContainsElement is a helper function to check if an element is in string array
func CheckStringContainsElement(stringItem string, splitChar string, find string) bool {
	s := strings.Split(stringItem, splitChar)
	for _, v := range s {
		if v == find {
			return true
		}
	}

	return false
}

// GetMapKeysAsStringSlice returns a string map's keys as a slice of strings
func GetMapKeysAsStringSlice(stringMap map[string]string) []string {
	var keys []string
	for k := range stringMap {
		keys = append(keys, k)
	}

	return keys
}
