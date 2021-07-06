package main

import (
	"io/ioutil"
	"os"
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
		{"65530", "PEER_65530"},
	}
	for _, tc := range testCases {
		if out := *sanitize(tc.input); out != tc.expectedOutput {
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

func TestMoveFile(t *testing.T) {
	// Make temporary cache directory
	if err := os.Mkdir("test-cache", 0755); err != nil && !os.IsExist(err) {
		t.Error(err)
	}

	inputString := "Test File"

	if err := ioutil.WriteFile("test-cache/source.txt", []byte(inputString), 0755); err != nil {
		t.Error(err)
	}

	if err := moveFile("test-cache/source.txt", "test-cache/dest.txt"); err != nil {
		t.Error(err)
	}

	if _, err := os.Stat("test-cache/dest.txt"); os.IsNotExist(err) {
		t.Errorf("file text-cache/dest.txt doesn't exist but should")
	}

	if _, err := os.Stat("test-cache/source.txt"); err == nil {
		t.Errorf("file text-cache/source.txt exists but shouldn't")
	}

	if contents, err := ioutil.ReadFile("test-cache/dest.txt"); err != nil {
		if string(contents) != inputString {
			t.Errorf("expected %s got %s", inputString, contents)
		}
	}

	if err := os.Remove("test-cache/dest.txt"); err != nil {
		t.Error(err)
	}
}
