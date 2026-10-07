package irr

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/natesales/pathvector/pkg/config"
	"github.com/natesales/pathvector/pkg/util"
)

func TestCacheRoundTrip(t *testing.T) {
	path := CachePath(t.TempDir(), 65510, "EXAMPLE")
	assert.Equal(t, "AS65510_EXAMPLE.json", filepath.Base(path))

	c := LoadCache(path) // missing file: empty cache
	assert.Nil(t, c.prefixes("k"))
	require.NoError(t, c.Save()) // nothing to save
	assert.NoFileExists(t, path)

	c.storePrefixes("k", []string{"192.0.2.0/24"}, []string{})
	c.storeMembers("m", []uint32{65510})
	require.NoError(t, c.Save())

	c = LoadCache(path)
	require.NotNil(t, c.prefixes("k"))
	assert.Equal(t, []string{"192.0.2.0/24"}, c.prefixes("k").PrefixSet4)
	assert.Nil(t, c.prefixes("other"), "data for a different query key must not be used")
	assert.Equal(t, []uint32{65510}, c.members("m").ASSetMembers)
	assert.Nil(t, c.members("other"))

	// Invalid file: empty cache
	require.NoError(t, os.WriteFile(path, []byte("{invalid"), 0600))
	assert.Nil(t, LoadCache(path).prefixes("k"))

	// nil cache is a no-op
	var nc *Cache
	nc.storePrefixes("k", nil, nil)
	assert.Nil(t, nc.prefixes("k"))
	assert.NoError(t, nc.Save())
}

func TestUpdateWithCacheFallback(t *testing.T) {
	path := CachePath(t.TempDir(), 65510, "EXAMPLE")
	newPeer := func() *config.Peer {
		return &config.Peer{ASSet: util.Ptr("AS-EXAMPLE"), NeighborIPs: &[]string{"192.0.2.1"}}
	}

	fakeBGPQ4(t, `case "$*" in *-Ab4*) printf 'NN = [\n    192.0.2.0/24\n];\n' ;; *) echo 'NN = [ ];' ;; esac`)
	c := LoadCache(path)
	require.NoError(t, UpdateWithCache(newPeer(), "rr.ntt.net", 5, "", c, false))
	require.NoError(t, c.Save())

	fakeBGPQ4(t, "exit 1")
	p := newPeer()
	require.NoError(t, UpdateWithCache(p, "rr.ntt.net", 5, "", LoadCache(path), false))
	assert.Equal(t, []string{"192.0.2.0/24"}, *p.PrefixSet4)

	// Different bgpq4 arguments are a different query
	assert.Error(t, UpdateWithCache(newPeer(), "rr.ntt.net", 5, "-S RADB", LoadCache(path), false))
	// No cache
	assert.Error(t, UpdateWithCache(newPeer(), "rr.ntt.net", 5, "", nil, false))
	// Offline with cache
	p = newPeer()
	require.NoError(t, UpdateWithCache(p, "rr.ntt.net", 5, "", LoadCache(path), true))
	assert.Equal(t, []string{"192.0.2.0/24"}, *p.PrefixSet4)
	// Members: offline without cache
	_, err := ASMembersWithCache("AS-EXAMPLE", "rr.ntt.net", 5, "", LoadCache(path), true)
	assert.ErrorIs(t, err, errOffline)
}
