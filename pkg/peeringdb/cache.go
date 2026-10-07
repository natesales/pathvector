package peeringdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/natesales/pathvector/pkg/util"
)

// This file implements a persistent on-disk cache of PeeringDB results. Unlike the in-memory cache in NetworkInfo
// (which only lives for one run), it is used as a fallback when PeeringDB can't be reached, so that configuration
// can still be generated from the last known data.

// errOffline is returned in place of a live query error when live PeeringDB queries are disabled
var errOffline = errors.New("offline mode, live PeeringDB queries disabled")

// errNoPage is wrapped by errors that mean PeeringDB answered and the network has no page. This is a definitive
// answer rather than a connectivity problem, so it doesn't trigger the fallback to cached data.
var errNoPage = errors.New("doesn't have a PeeringDB page")

// cacheEntry is the on-disk format of a cached PeeringDB result
type cacheEntry[T any] struct {
	Updated time.Time `json:"updated"`
	Data    T         `json:"data"`
}

// networkCachePath returns the cache file path for an ASN's network info
func networkCachePath(cacheDirectory string, asn uint32) string {
	return filepath.Join(cacheDirectory, "peeringdb", fmt.Sprintf("AS%d.json", asn))
}

// nvrsCachePath returns the cache file path for the never via route servers list
func nvrsCachePath(cacheDirectory string) string {
	return filepath.Join(cacheDirectory, "peeringdb", "never-via-route-servers.json")
}

// writeCache stores data in a cache file, logging (but otherwise ignoring) failures
func writeCache[T any](path string, data T) {
	b, err := json.MarshalIndent(cacheEntry[T]{Updated: time.Now(), Data: data}, "", "  ")
	if err == nil {
		err = util.WriteFileAtomic(path, b)
	}
	if err != nil {
		log.Warnf("Writing PeeringDB cache %s: %v", path, err)
	}
}

// readCache reads a cache file
func readCache[T any](path string) (*cacheEntry[T], error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entry cacheEntry[T]
	if err := json.Unmarshal(b, &entry); err != nil {
		return nil, fmt.Errorf("invalid cache file %s: %w", path, err)
	}
	return &entry, nil
}

// withFallback runs a live query (unless offline) and stores its result in the cache file at path. If the live query
// fails with anything other than errNoPage, the cached result is returned instead with a warning.
// An empty cacheDirectory disables the on-disk cache.
func withFallback[T any](what, cacheDirectory, path string, offline bool, live func() (T, error)) (T, error) {
	var result T
	err := errOffline
	if !offline {
		result, err = live()
	}
	if cacheDirectory == "" {
		return result, err
	}
	if err == nil {
		writeCache(path, result)
		return result, nil
	}
	if errors.Is(err, errNoPage) {
		return result, err
	}

	cached, cacheErr := readCache[T](path)
	if cacheErr != nil {
		if !os.IsNotExist(cacheErr) {
			log.Warnf("Reading PeeringDB cache: %v", cacheErr)
		}
		return result, err
	}
	log.Warnf("%s: %v; using cached PeeringDB data from %s (%s old)", what, err, cached.Updated.Format(time.RFC3339), time.Since(cached.Updated).Round(time.Second))
	return cached.Data, nil
}

// NetworkInfoWithFallback gets the PeeringDB info for an ASN like NetworkInfo, and additionally stores successful
// results in cacheDirectory, falling back to the stored result if PeeringDB can't be queried (or offline is true).
func NetworkInfoWithFallback(asn uint32, queryTimeout uint, apiKey string, useCache bool, cacheDirectory string, offline bool) (*Data, error) {
	return withFallback(fmt.Sprintf("unable to get PeeringDB data for AS%d", asn), cacheDirectory, networkCachePath(cacheDirectory, asn), offline, func() (*Data, error) {
		return NetworkInfo(asn, queryTimeout, apiKey, useCache)
	})
}

// NeverViaRouteServersWithFallback gets the never via route servers list like NeverViaRouteServers, and additionally
// stores successful results in cacheDirectory, falling back to the stored result if PeeringDB can't be queried
// (or offline is true).
func NeverViaRouteServersWithFallback(queryTimeout uint, apiKey string, cacheDirectory string, offline bool) ([]uint32, error) {
	return withFallback("unable to get PeeringDB never via route servers list", cacheDirectory, nvrsCachePath(cacheDirectory), offline, func() ([]uint32, error) {
		return NeverViaRouteServers(queryTimeout, apiKey)
	})
}
