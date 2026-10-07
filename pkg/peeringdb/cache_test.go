package peeringdb

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/natesales/pathvector/pkg/config"
	"github.com/natesales/pathvector/pkg/util"
)

// fakePeeringDB serves a minimal PeeringDB API for AS65510 and the NVRS list
func fakePeeringDB(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Query().Get("asn") == "65510":
			_, _ = w.Write([]byte(`{"data": [{"name": "Example", "asn": 65510, "irr_as_set": "RIPE::AS-EXAMPLE", "info_prefixes4": 10, "info_prefixes6": 20}]}`))
		case r.URL.Query().Get("info_never_via_route_servers") == "1":
			_, _ = w.Write([]byte(`{"data": [{"asn": 174}, {"asn": 3356}]}`))
		default:
			_, _ = w.Write([]byte(`{"data": []}`))
		}
	}))
	orig := Endpoint
	Endpoint = srv.URL
	t.Cleanup(func() {
		srv.Close()
		Endpoint = orig
	})
	return srv
}

func TestNetworkInfoWithFallback(t *testing.T) {
	cacheDir := t.TempDir()
	srv := fakePeeringDB(t)

	// Live query succeeds and is cached on disk
	d, err := NetworkInfoWithFallback(65510, 5, "", false, cacheDir, false)
	require.NoError(t, err)
	assert.Equal(t, "RIPE::AS-EXAMPLE", d.ASSet)
	assert.FileExists(t, filepath.Join(cacheDir, "peeringdb", "AS65510.json"))

	// No PeeringDB page is a definitive answer and doesn't fall back
	_, err = NetworkInfoWithFallback(65520, 5, "", false, cacheDir, false)
	assert.ErrorIs(t, err, errNoPage)

	// PeeringDB unreachable: fall back to the cached data
	srv.Close()
	d, err = NetworkInfoWithFallback(65510, 1, "", false, cacheDir, false)
	require.NoError(t, err)
	assert.Equal(t, "RIPE::AS-EXAMPLE", d.ASSet)
	assert.Equal(t, 10, d.ImportLimit4)
	assert.Equal(t, 20, d.ImportLimit6)

	// Unreachable and nothing cached: error
	_, err = NetworkInfoWithFallback(65530, 1, "", false, cacheDir, false)
	assert.Error(t, err)

	// Offline mode uses only the cache
	d, err = NetworkInfoWithFallback(65510, 1, "", false, cacheDir, true)
	require.NoError(t, err)
	assert.Equal(t, "RIPE::AS-EXAMPLE", d.ASSet)
	_, err = NetworkInfoWithFallback(65530, 1, "", false, cacheDir, true)
	assert.ErrorIs(t, err, errOffline)

	// Empty cache directory disables the disk cache
	_, err = NetworkInfoWithFallback(65510, 1, "", false, "", false)
	assert.Error(t, err)
}

func TestNeverViaRouteServersWithFallback(t *testing.T) {
	cacheDir := t.TempDir()
	srv := fakePeeringDB(t)

	asns, err := NeverViaRouteServersWithFallback(5, "", cacheDir, false)
	require.NoError(t, err)
	assert.Equal(t, []uint32{174, 3356}, asns)

	srv.Close()
	asns, err = NeverViaRouteServersWithFallback(1, "", cacheDir, false)
	require.NoError(t, err)
	assert.Equal(t, []uint32{174, 3356}, asns)

	asns, err = NeverViaRouteServersWithFallback(1, "", cacheDir, true)
	require.NoError(t, err)
	assert.Equal(t, []uint32{174, 3356}, asns)
}

func TestUpdateFromDataDoesNotModifySource(t *testing.T) {
	d := &Data{ASN: 65510}
	peer := &config.Peer{ASN: util.Ptr(65510), AutoImportLimits: util.Ptr(false), AutoASSet: util.Ptr(true)}
	UpdateFromData(peer, d)
	assert.Equal(t, "AS65510", *peer.ASSet)
	assert.Equal(t, "", d.ASSet)
}
