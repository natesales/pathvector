package main

import (
	"testing"
)

func TestContains(t *testing.T) {
	testCases := []struct {
		array          []string
		element        string
		expectedOutput bool
	}{
		{[]string{"foo", "bar"}, "foo", true},
		{[]string{"foo", "bar"}, "baz", false},
	}
	for _, tc := range testCases {
		if out := contains(tc.array, tc.element); out != tc.expectedOutput {
			t.Errorf("array %+v element %s failed. expected '%v' got '%v'", tc.array, tc.element, tc.expectedOutput, out)
		}
	}
}

func TestSanitize(t *testing.T) {
	testCases := []struct {
		input          string
		expectedOutput string
	}{
		{"foo", "FOO"},
		{"fooBAR", "FOOBAR"},
		{"fooBAR---", "FOOBAR"},
		{"fooBAR-*-", "FOOBAR"},
		{"FOOBAR", "FOOBAR"},
	}
	for _, tc := range testCases {
		if out := sanitize(tc.input); out != tc.expectedOutput {
			t.Errorf("sanitize %s failed. expected '%v' got '%v'", tc.input, tc.expectedOutput, out)
		}
	}
}
