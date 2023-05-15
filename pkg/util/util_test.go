package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
		if out := Contains(tc.array, tc.element); out != tc.expectedOutput {
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
		if out := *Sanitize(tc.input); out != tc.expectedOutput {
			t.Errorf("sanitize %s failed. expected '%v' got '%v'", tc.input, tc.expectedOutput, out)
		}
	}
}

func TestMoveFile(t *testing.T) {
	// Make temporary cache directory
	if err := os.Mkdir("test-cache", 0755); err != nil && !os.IsExist(err) {
		t.Error(err)
	}

	inputString := "Test File"

	//nolint:golint,gosec
	if err := os.WriteFile("test-cache/source.txt", []byte(inputString), 0644); err != nil {
		t.Error(err)
	}

	if err := MoveFile("test-cache/source.txt", "test-cache/dest.txt"); err != nil {
		t.Error(err)
	}

	if _, err := os.Stat("test-cache/dest.txt"); os.IsNotExist(err) {
		t.Errorf("file text-cache/dest.txt doesn't exist but should")
	}

	if _, err := os.Stat("test-cache/source.txt"); err == nil {
		t.Errorf("file text-cache/source.txt exists but shouldn't")
	}

	if contents, err := os.ReadFile("test-cache/dest.txt"); err != nil {
		if string(contents) != inputString {
			t.Errorf("expected %s got %s", inputString, contents)
		}
	}

	if err := os.Remove("test-cache/dest.txt"); err != nil {
		t.Error(err)
	}
}

func TestPrintTable(t *testing.T) {
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	PrintTable([]string{"foo", "bar", "baz"}, [][]string{{"foo", "bar", "baz"}, {"foo", "bar", "baz"}})
	w.Close()
	os.Stdout = old
}

func TestUtilPtrDeref(t *testing.T) {
	assert.Equal(t, true, Deref(Ptr(true)))
	assert.Equal(t, false, Deref(Ptr(false)))
	assert.Equal(t, false, Deref[bool](nil))

	assert.Equal(t, "foo", StrDeref(Ptr("foo")))
	assert.Equal(t, "", StrDeref(nil))
}
