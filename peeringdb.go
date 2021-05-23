package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

// peeringDbResponse contains the response from a PeeringDB query
type peeringDbResponse struct {
	Data []peeringDbData `json:"data"`
}

// peeringDbData contains the actual data from PeeringDB response
type peeringDbData struct {
	Name    string `json:"name"`
	AsSet   string `json:"irr_as_set"`
	MaxPfx4 uint   `json:"info_prefixes4"`
	MaxPfx6 uint   `json:"info_prefixes6"`
}

// Query PeeringDB for an ASN
func getPeeringDbData(asn uint) peeringDbData {
	httpClient := http.Client{Timeout: time.Second * 5}
	req, err := http.NewRequest(http.MethodGet, "https://peeringdb.com/api/net?asn="+strconv.Itoa(int(asn)), nil)
	if err != nil {
		log.Fatalf("PeeringDB GET (This peer might not have a PeeringDB page): %v", err)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		log.Fatalf("PeeringDB GET Request: %v", err)
	}

	if res.Body != nil {
		//noinspection GoUnhandledErrorResult
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("PeeringDB Read: %v", err)
	}

	var pDbResponse peeringDbResponse
	if err := json.Unmarshal(body, &pDbResponse); err != nil {
		log.Fatalf("PeeringDB JSON Unmarshal: %v", err)
	}

	if len(pDbResponse.Data) < 1 {
		log.Fatalf("Peer %d doesn't have a PeeringDB entry.", asn)
	}

	return pDbResponse.Data[0]
}
