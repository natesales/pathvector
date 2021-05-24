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
