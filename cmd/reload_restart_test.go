package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCtlReloadResetParseArgs(t *testing.T) {
	for _, tc := range []struct {
		args         []string
		expQuery     string
		expDirection string
	}{
		{[]string{"in", "all"}, "all", "in"},
		{[]string{"out", "all"}, "all", "out"},
		{[]string{"all"}, "all", "both"},
		{[]string{"in", "he"}, "he", "in"},
		{[]string{"out", "he"}, "he", "out"},
		{[]string{"he"}, "he", "both"},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			query, direction := parseArgs(tc.args)
			assert.Equal(t, tc.expQuery, query)
			assert.Equal(t, tc.expDirection, direction)
		})
	}
}
