package peeringdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/natesales/pathvector/pkg/config"
)

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

var apiKey = ""

func InitAPIKey(key string) {
	if key != "" {
		apiKey = key
	}
}

// NetworkInfo returns PeeringDB for an ASN
func NetworkInfo(asn uint, queryTimeout uint) (*Data, error) {
	httpClient := http.Client{Timeout: time.Second * time.Duration(queryTimeout)}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://peeringdb.com/api/net?asn=%d", asn), nil)

	if apiKey != "" {
		req.Header.Add("AUTHORIZATION", fmt.Sprintf("Api-Key %s", apiKey))
	}

	if err != nil {
		return nil, errors.New("PeeringDB GET (This peer might not have a PeeringDB page): " + err.Error())
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, errors.New("PeeringDB GET request: " + err.Error())
	}

	if res.Body != nil {
		//noinspection GoUnhandledErrorResult
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.New("PeeringDB read: " + err.Error())
	}

	var pDbResponse Response
	if err := json.Unmarshal(body, &pDbResponse); err != nil {
		return nil, errors.New("PeeringDB JSON Unmarshal: " + err.Error())
	}

	if len(pDbResponse.Data) < 1 {
		return nil, fmt.Errorf("peer %d doesn't have a PeeringDB page", asn)
	}

	return &pDbResponse.Data[0], nil // nil error
}

// Update updates peer values from PeeringDB
func Update(peerData *config.Peer, queryTimeout uint) {
	pDbData, err := NetworkInfo(uint(*peerData.ASN), queryTimeout)
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

		// Used to get address of string
		asSetOutput := sanitizeASSet(pDbData.ASSet)
		peerData.ASSet = &asSetOutput
	}
}

// NeverViaRouteServers gets a list of networks that report should never be reachable via route servers
func NeverViaRouteServers(queryTimeout uint) ([]uint32, error) {
	httpClient := http.Client{Timeout: time.Second * time.Duration(queryTimeout)}
	req, err := http.NewRequest(http.MethodGet, "https://peeringdb.com/api/net?info_never_via_route_servers=1", nil)
	if err != nil {
		return nil, errors.New("PeeringDB GET: " + err.Error())
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, errors.New("PeeringDB GET request: " + err.Error())
	}
	if res.Body != nil {
		//noinspection GoUnhandledErrorResult
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.New("PeeringDB read: " + err.Error())
	}

	var pDbResponse Response
	if err := json.Unmarshal(body, &pDbResponse); err != nil {
		return nil, errors.New("PeeringDB JSON Unmarshal: " + err.Error())
	}

	var asns []uint32 // ASNs that are reportedly never reachable via route servers
	for _, resp := range pDbResponse.Data {
		asns = append(asns, resp.ASN)
	}

	return asns, nil // nil error
}

// sanitizeASSet removes an IRRDB prefix from an AS set and picks the first one if there are multiple
func sanitizeASSet(asSet string) string {
	output := asSet

	// If the as-set has a space in it, split and pick the first one
	if strings.Contains(output, " ") {
		output = strings.Split(output, " ")[0]
		log.Warnf("Original as-set %s has a space in it. Selecting first element %s", asSet, output)
	}

	// Trim IRRDB prefix
	if strings.Contains(output, "::") {
		output = strings.Split(output, "::")[1]
		log.Warnf("Original as-set %s has an IRRDB prefix in it. Using %s", asSet, output)
	}

	return output
}
