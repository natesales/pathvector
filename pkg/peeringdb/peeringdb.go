package peeringdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/natesales/pathvector/pkg/config"
)

// Endpoint is a public value to allow setting to a cache server
var Endpoint = ""

func init() {
	// Check if running in test
	if os.Getenv("PATHVECTOR_TEST") == "1" {
		Endpoint = "http://localhost:5000/api"
	}
}

type IxLanResponse struct {
	Data []IxLanData `json:"data"`
}

type IxLanData struct {
	Id          int       `json:"id"`
	NetId       int       `json:"net_id"`
	IxId        int       `json:"ix_id"`
	Name        string    `json:"name"`
	IxlanId     int       `json:"ixlan_id"`
	Notes       string    `json:"notes"`
	Speed       int       `json:"speed"`
	Asn         uint32    `json:"asn"`
	Ipaddr4     string    `json:"ipaddr4"`
	Ipaddr6     string    `json:"ipaddr6"`
	IsRsPeer    bool      `json:"is_rs_peer"`
	Operational bool      `json:"operational"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Status      string    `json:"status"`
}

// Response contains the response from a PeeringDB query
type Response struct {
	Data []Data `json:"data"`
}

// Data contains the actual data from PeeringDB response
type Data struct {
	Name         string `json:"name"`
	ASN          uint32 `json:"asn"`
	ASSet        string `json:"irr_as_set"`
	ImportLimit4 int    `json:"info_prefixes4"`
	ImportLimit6 int    `json:"info_prefixes6"`
}

var (
	cache map[uint32]*Data
	lock  sync.Mutex
)

// networkInfo returns PeeringDB for an ASN
func networkInfo(asn uint32, queryTimeout uint, apiKey string) (*Data, error) {
	httpClient := http.Client{Timeout: time.Second * time.Duration(queryTimeout)}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(Endpoint+"/net?asn=%d", asn), nil)

	if apiKey != "" {
		req.Header.Add("AUTHORIZATION", "Api-Key "+apiKey)
	} else if os.Getenv("PEERINGDB_API_KEY") != "" {
		req.Header.Add("AUTHORIZATION", "Api-Key "+os.Getenv("PEERINGDB_API_KEY"))
	}

	if err != nil {
		return nil, fmt.Errorf("PeeringDB GET (This peer might not have a PeeringDB page): %s", err)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("PeeringDB GET request: %s", err)
	}

	if res.StatusCode == 404 {
		return nil, fmt.Errorf("peer %d doesn't have a PeeringDB page", asn)
	}

	if res.StatusCode != 200 {
		return nil, errors.New("PeeringDB GET request expected 200, got " + res.Status)
	}

	if res.Body != nil {
		//noinspection GoUnhandledErrorResult
		defer res.Body.Close()
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("PeeringDB read: %s", err)
	}

	var pDbResponse Response
	if err := json.Unmarshal(body, &pDbResponse); err != nil {
		return nil, fmt.Errorf("%s PeeringDB JSON Unmarshal: %s", req.URL, err)
	}

	if len(pDbResponse.Data) < 1 {
		return nil, fmt.Errorf("peer %d doesn't have a PeeringDB page", asn)
	}

	return &pDbResponse.Data[0], nil // nil error
}

// NetworkInfo gets the PeeringDB info for an ASN optionally from the cache
func NetworkInfo(asn uint32, queryTimeout uint, apiKey string, useCache bool) (*Data, error) {
	if !useCache {
		return networkInfo(asn, queryTimeout, apiKey)
	} else {
		lock.Lock()
		defer lock.Unlock()
		if cache == nil {
			cache = make(map[uint32]*Data)
		}
		if _, ok := cache[asn]; !ok {
			d, err := networkInfo(asn, queryTimeout, apiKey)
			if err != nil {
				return nil, err
			}
			cache[asn] = d
		}

		return cache[asn], nil
	}
}

// Update updates peer values from PeeringDB
func Update(peerData *config.Peer, queryTimeout uint, apiKey string, useCache bool) {
	pDbData, err := NetworkInfo(uint32(*peerData.ASN), queryTimeout, apiKey, useCache)
	if err != nil {
		log.Fatalf("unable to get PeeringDB data: %+v", err)
	}

	// Set import limits
	if *peerData.AutoImportLimits {
		peerData.ImportLimit4 = &pDbData.ImportLimit4
		peerData.ImportLimit6 = &pDbData.ImportLimit6

		if pDbData.ImportLimit4 == 0 {
			log.Warnf("peer AS%d has an IPv4 import limit of zero from PeeringDB", *peerData.ASN)
		}
		if pDbData.ImportLimit6 == 0 {
			log.Warnf("peer AS%d has an IPv6 import limit of zero from PeeringDB", *peerData.ASN)
		}
	}

	// Set as-set if auto-as-set is enabled and there isn't a manual AS set defined
	if *peerData.AutoASSet && peerData.ASSet == nil {
		if pDbData.ASSet == "" {
			log.Warnf("peer AS%d doesn't have an as-set in PeeringDB, using ASN instead", *peerData.ASN)
			pDbData.ASSet = fmt.Sprintf("AS%d", *peerData.ASN)
		}

		peerData.ASSet = &pDbData.ASSet
	}
}

// NeverViaRouteServers gets a list of networks that report should never be reachable via route servers
func NeverViaRouteServers(queryTimeout uint, apiKey string) ([]uint32, error) {
	httpClient := http.Client{Timeout: time.Second * time.Duration(queryTimeout)}
	req, err := http.NewRequest(http.MethodGet, Endpoint+"/net?info_never_via_route_servers=1", nil)
	if err != nil {
		return nil, fmt.Errorf("PeeringDB GET: %s", err)
	}

	if apiKey != "" {
		req.Header.Add("AUTHORIZATION", "Api-Key "+apiKey)
	} else {
		if os.Getenv("PEERINGDB_API_KEY") != "" {
			req.Header.Add("AUTHORIZATION", "Api-Key "+os.Getenv("PEERINGDB_API_KEY"))
		}
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("PeeringDB GET request: %s", err)
	}
	if res.StatusCode != 200 {
		return nil, errors.New("PeeringDB GET request expected 200, got " + res.Status)
	}
	if res.Body != nil {
		//noinspection GoUnhandledErrorResult
		defer res.Body.Close()
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("PeeringDB read: %s", err)
	}

	var pDbResponse Response
	if err := json.Unmarshal(body, &pDbResponse); err != nil {
		return nil, fmt.Errorf("%s PeeringDB JSON Unmarshal: %s", req.URL, err)
	}

	var asns []uint32 // ASNs that are reportedly never reachable via route servers
	for _, resp := range pDbResponse.Data {
		asns = append(asns, resp.ASN)
	}

	return asns, nil // nil error
}

// IXLANs gets PeeringDB IX LANs for an ASN
func IXLANs(asn uint32, peeringDbQueryTimeout uint, apiKey string) ([]IxLanData, error) {
	httpClient := http.Client{Timeout: time.Second * time.Duration(peeringDbQueryTimeout)}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(Endpoint+"/netixlan?asn=%d", asn), nil)
	if err != nil {
		return nil, fmt.Errorf("PeeringDB GET (This peer might not have a PeeringDB page): %s", err)
	}

	if apiKey != "" {
		req.Header.Add("AUTHORIZATION", "Api-Key "+apiKey)
	} else {
		if os.Getenv("PEERINGDB_API_KEY") != "" {
			req.Header.Add("AUTHORIZATION", "Api-Key "+os.Getenv("PEERINGDB_API_KEY"))
		}
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("PeeringDB GET request: %s", err)
	}
	if res.StatusCode != 200 {
		return nil, errors.New("PeeringDB GET request expected 200, got " + res.Status)
	}
	if res.Body != nil {
		//noinspection GoUnhandledErrorResult
		defer res.Body.Close()
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("PeeringDB read: %s", err)
	}

	var pDbResponse IxLanResponse
	if err := json.Unmarshal(body, &pDbResponse); err != nil {
		return nil, fmt.Errorf("%s PeeringDB JSON Unmarshal: %s", req.URL, err)
	}

	if len(pDbResponse.Data) < 1 {
		return nil, fmt.Errorf("peer %d doesn't have a PeeringDB page or IXPs documented", asn)
	}

	return pDbResponse.Data, nil // nil error
}
