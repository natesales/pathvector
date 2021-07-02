package peeringdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/natesales/pathvector/internal/config"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Response contains the response from a PeeringDB query
type Response struct {
	Data []Data `json:"data"`
}

// Data contains the actual data from PeeringDB response
type Data struct {
	Name         string `json:"name"`
	ASSet        string `json:"irr_as_set"`
	ImportLimit4 int    `json:"info_prefixes4"`
	ImportLimit6 int    `json:"info_prefixes6"`
}

// Query PeeringDB for an ASN
func getPeeringDbData(asn int, timeout uint) (*Data, error) {
	httpClient := http.Client{Timeout: time.Second * time.Duration(timeout)}
	req, err := http.NewRequest(http.MethodGet, "https://peeringdb.com/api/net?asn="+strconv.Itoa(int(asn)), nil)
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

// Run updates peer values from PeeringDB
func Run(peerData *config.Peer, timeout uint) error {
	pDbData, err := getPeeringDbData(*peerData.ASN, timeout)
	if err != nil {
		return fmt.Errorf("unable to get PeeringDB data: %+v", err)
	}

	// Set import limits
	if *peerData.AutoImportLimits {
		*peerData.ImportLimit6 = pDbData.ImportLimit4
		*peerData.ImportLimit6 = pDbData.ImportLimit6

		if pDbData.ImportLimit4 == 0 {
			return fmt.Errorf("peer has an IPv4 import limit of zero from PeeringDB")
		}
		if pDbData.ImportLimit6 == 0 {
			return fmt.Errorf("peer has an IPv6 import limit of zero from PeeringDB")
		}
	}

	// Set as-set if auto-as-set is enabled and there isn't a manual AS set defined
	if *peerData.AutoASSet && peerData.ASSet == nil {
		if pDbData.ASSet == "" {
			return fmt.Errorf("peer doesn't have an as-set in PeeringDB")
			// TODO: Exit or skip this peer?
		}

		// If the as-set has a space in it, split and pick the first one
		if strings.Contains(pDbData.ASSet, " ") {
			pDbData.ASSet = strings.Split(pDbData.ASSet, " ")[0]
			return fmt.Errorf("peer has a space in their PeeringDB as-set field. Selecting first element %s", pDbData.ASSet)
		}

		// Trim IRRDB prefix
		if strings.Contains(pDbData.ASSet, "::") {
			peerData.ASSet = &strings.Split(pDbData.ASSet, "::")[1]
			return fmt.Errorf("peer has an IRRDB prefix in their PeeringDB as-set field. Using %s", *peerData.ASSet)
		} else {
			peerData.ASSet = &pDbData.ASSet
		}
	}

	return nil // nil error
}
