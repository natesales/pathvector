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
		{"AS65530", "AS65530"},
		{"65530", "65530"},
	}
	for _, tc := range testCases {
		if out := sanitize(tc.input); out != tc.expectedOutput {
			t.Errorf("sanitize %s failed. expected '%v' got '%v'", tc.input, tc.expectedOutput, out)
		}
	}
}

func TestCategorizeCommunity(t *testing.T) {
	testCases := []struct {
		input          string
		expectedOutput string
		shouldError    bool
	}{
		{"34553,0", "standard", false},
		{"4242424242:4242424242:0", "large", false},
		{":", "", true},
		{"4242424242,0", "", true},
	}
	for _, tc := range testCases {
		cType := categorizeCommunity(tc.input)
		if cType != "" && tc.shouldError {
			t.Errorf("categorizeCommunity should have errored on '%s' but didn't. expected error, got '%s'", tc.input, cType)
		} else if cType == "" && !tc.shouldError {
			t.Errorf("categorizeCommunity shouldn't have errored on '%s' but did. expected '%s'", tc.input, tc.expectedOutput)
		} else if cType != tc.expectedOutput {
			t.Errorf("categorizeCommunity %s failed. expected '%v' got '%v'", tc.input, tc.expectedOutput, cType)
		}
	}
}
