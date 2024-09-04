package util

import (
	"os"
	"slices"
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
		if out := slices.Contains(tc.array, tc.element); out != tc.expectedOutput {
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
	assert.Nil(t, os.WriteFile("test-cache/source.txt", []byte(inputString), 0644))

	assert.Nil(t, MoveFile("test-cache/source.txt", "test-cache/dest.txt"))

	if _, err := os.Stat("test-cache/dest.txt"); os.IsNotExist(err) {
		t.Errorf("file text-cache/dest.txt doesn't exist but should")
	}

	if _, err := os.Stat("test-cache/source.txt"); err == nil {
		t.Errorf("file text-cache/source.txt exists but shouldn't")
	}

	contents, err := os.ReadFile("test-cache/dest.txt")
	assert.Nil(t, err)
	assert.Equal(t, inputString, string(contents))

	assert.Nil(t, os.Remove("test-cache/dest.txt"))
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

func TestUtilIsPrivateASN(t *testing.T) {
	assert.True(t, IsPrivateASN(65534))
	assert.True(t, IsPrivateASN(65535))
	assert.True(t, IsPrivateASN(4200000000))
	assert.False(t, IsPrivateASN(112))
}
