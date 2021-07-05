package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// peeringDbResponse contains the response from a PeeringDB query
type peeringDbResponse struct {
	Data []peeringDbData `json:"data"`
}

// peeringDbData contains the actual data from PeeringDB response
type peeringDbData struct {
	Name         string `json:"name"`
	ASSet        string `json:"irr_as_set"`
	ImportLimit4 int    `json:"info_prefixes4"`
	ImportLimit6 int    `json:"info_prefixes6"`
}

// Query PeeringDB for an ASN
func getPeeringDbData(asn int) (*peeringDbData, error) {
	httpClient := http.Client{Timeout: time.Second * time.Duration(peeringDbQueryTimeout)}
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

	var pDbResponse peeringDbResponse
	if err := json.Unmarshal(body, &pDbResponse); err != nil {
		return nil, errors.New("PeeringDB JSON Unmarshal: " + err.Error())
	}

	if len(pDbResponse.Data) < 1 {
		return nil, fmt.Errorf("peer %d doesn't have a PeeringDB page", asn)
	}

	return &pDbResponse.Data[0], nil // nil error
}

// runPeeringDbQuery updates peer values from PeeringDB
func runPeeringDbQuery(peerData *Peer) error {
	pDbData, err := getPeeringDbData(*peerData.ASN)
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

		// Used to get address of string
		asSetOutput := sanitizeASSet(pDbData.ASSet)
		peerData.ASSet = &asSetOutput
	}

	return nil // nil error
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
