package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// peeringDbResponse contains the response from a PeeringDB query
type peeringDbResponse struct {
	Data []peeringDbData `json:"data"`
}

// peeringDbData contains the actual data from PeeringDB response
type peeringDbData struct {
	Name         string `json:"name"`
	ASSet        string `json:"irr_as_set"`
	ImportLimit4 uint   `json:"info_prefixes4"`
	ImportLimit6 uint   `json:"info_prefixes6"`
}

// Query PeeringDB for an ASN
func getPeeringDbData(asn uint) (*peeringDbData, error) {
	httpClient := http.Client{Timeout: time.Second * time.Duration(cliFlags.PeeringDbQueryTimeout)}
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
		return nil, errors.New(fmt.Sprintf("peer %d doesn't have a PeeringDB page", asn))
	}

	return &pDbResponse.Data[0], nil // nil error
}
