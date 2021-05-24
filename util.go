package main

import "strings"

var alphabet = strings.Split("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", "")

// contains runs a linear search on a string array
func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// sanitize limits an input string to only uppercase letters and numbers
func sanitize(input string) string {
	output := ""
	for _, chr := range []rune(strings.ToUpper(input)) {
		if contains(alphabet, string(chr)) {
			output += string(chr)
		}
	}
	return output
}
