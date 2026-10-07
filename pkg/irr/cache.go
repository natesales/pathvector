package irr

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

// errOffline is returned in place of a live query error when live IRR queries are disabled
var errOffline = errors.New("offline mode, live IRR queries disabled")

// CachedPrefixes holds the IRR prefix sets from the last successful query for a peer
type CachedPrefixes struct {
	// Key identifies the query (as-set and bgpq4 arguments) that produced this data
	Key        string    `json:"key"`
	Updated    time.Time `json:"updated"`
	PrefixSet4 []string  `json:"prefix_set4"`
	PrefixSet6 []string  `json:"prefix_set6"`
}

// CachedMembers holds the AS set members from the last successful query for a peer
type CachedMembers struct {
	// Key identifies the query (as-set and bgpq4 arguments) that produced this data
	Key          string    `json:"key"`
	Updated      time.Time `json:"updated"`
	ASSetMembers []uint32  `json:"as_set_members"`
}

// cacheData is the on-disk format of a peer's IRR cache file
type cacheData struct {
	Prefixes *CachedPrefixes `json:"prefixes,omitempty"`
	Members  *CachedMembers  `json:"members,omitempty"`
}

// Cache is a persistent per-peer cache of IRR query results, used as a fallback when live IRR queries fail.
// A nil *Cache is valid and disables caching. A Cache is not safe for concurrent use; each peer has its own.
type Cache struct {
	path  string
	data  cacheData
	dirty bool
}

// CachePath returns the IRR cache file path for a peer
func CachePath(cacheDirectory string, asn int, sanitizedName string) string {
	return filepath.Join(cacheDirectory, "irr", fmt.Sprintf("AS%d_%s.json", asn, sanitizedName))
}

// LoadCache loads an IRR cache file. A missing or unreadable file results in an empty cache.
func LoadCache(path string) *Cache {
	c := &Cache{path: path}
	b, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Warnf("Reading IRR cache %s: %v", path, err)
		}
		return c
	}
	if err := json.Unmarshal(b, &c.data); err != nil {
		log.Warnf("Ignoring invalid IRR cache %s: %v", path, err)
		c.data = cacheData{}
	}
	return c
}

// Save writes the cache to disk if it has changed
func (c *Cache) Save() error {
	if c == nil || !c.dirty {
		return nil
	}
	b, err := json.MarshalIndent(c.data, "", "  ")
	if err != nil {
		return err
	}
	if err := util.WriteFileAtomic(c.path, b); err != nil {
		return fmt.Errorf("writing IRR cache %s: %w", c.path, err)
	}
	c.dirty = false
	return nil
}

func (c *Cache) storePrefixes(key string, pfx4, pfx6 []string) {
	if c == nil {
		return
	}
	c.data.Prefixes = &CachedPrefixes{Key: key, Updated: time.Now(), PrefixSet4: pfx4, PrefixSet6: pfx6}
	c.dirty = true
}

// prefixes returns the cached prefix sets for a query key, or nil if there are none
func (c *Cache) prefixes(key string) *CachedPrefixes {
	if c == nil || c.data.Prefixes == nil {
		return nil
	}
	if c.data.Prefixes.Key != key {
		log.Debugf("Ignoring cached IRR prefix sets in %s for a different query (%s, want %s)", c.path, c.data.Prefixes.Key, key)
		return nil
	}
	return c.data.Prefixes
}

func (c *Cache) storeMembers(key string, members []uint32) {
	if c == nil {
		return
	}
	c.data.Members = &CachedMembers{Key: key, Updated: time.Now(), ASSetMembers: members}
	c.dirty = true
}

// members returns the cached AS set members for a query key, or nil if there are none
func (c *Cache) members(key string) *CachedMembers {
	if c == nil || c.data.Members == nil {
		return nil
	}
	if c.data.Members.Key != key {
		log.Debugf("Ignoring cached AS set members in %s for a different query (%s, want %s)", c.path, c.data.Members.Key, key)
		return nil
	}
	return c.data.Members
}
