package match

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/natesales/pathvector/internal/peeringdb"
)

type peeringDbIxLanResponse struct {
	Data []peeringDbIxLanData `json:"data"`
}

type peeringDbIxLanData struct {
	Id          int       `json:"id"`
	NetId       int       `json:"net_id"`
	IxId        int       `json:"ix_id"`
	Name        string    `json:"name"`
	IxlanId     int       `json:"ixlan_id"`
	Notes       string    `json:"notes"`
	Speed       int       `json:"speed"`
	Asn         int       `json:"asn"`
	Ipaddr4     string    `json:"ipaddr4"`
	Ipaddr6     string    `json:"ipaddr6"`
	IsRsPeer    bool      `json:"is_rs_peer"`
	Operational bool      `json:"operational"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Status      string    `json:"status"`
}

func peeringDbIxLans(asn uint, peeringDbQueryTimeout uint) ([]peeringDbIxLanData, error) {
	httpClient := http.Client{Timeout: time.Second * time.Duration(peeringDbQueryTimeout)}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://peeringdb.com/api/netixlan?asn=%d", asn), nil)
	if err != nil {
		return nil, fmt.Errorf("PeeringDB GET (This peer might not have a PeeringDB page): %s", err)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("PeeringDB GET request: %s", err)
	}

	if res.Body != nil {
		//noinspection GoUnhandledErrorResult
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("PeeringDB read: %s", err)
	}

	var pDbResponse peeringDbIxLanResponse
	if err := json.Unmarshal(body, &pDbResponse); err != nil {
		return nil, fmt.Errorf("PeeringDB JSON Unmarshal: %s", err)
	}

	if len(pDbResponse.Data) < 1 {
		return nil, fmt.Errorf("peer %d doesn't have a PeeringDB page or IXPs documented", asn)
	}

	return pDbResponse.Data, nil // nil error
}

// CommonIXs gets common IXPs from PeeringDB
func CommonIXs(a uint, b uint, yamlFormat bool, queryTimeout uint, apiKey string) string {
	networkA, err := peeringDbIxLans(a, queryTimeout)
	if err != nil {
		log.Fatalf("AS%d: %v", a, err)
	}
	networkB, err := peeringDbIxLans(b, queryTimeout)
	if err != nil {
		log.Fatalf("AS%d: %v", a, err)
	}

	networkBInfo, err := peeringdb.NetworkInfo(b, queryTimeout, apiKey)
	if err != nil {
		log.Fatalf("AS%d: %v", b, err)
	}

	out := ""
	for _, ixA := range networkA {
		for _, ixB := range networkB {
			if ixA.IxlanId == ixB.IxlanId {
				if !yamlFormat {
					out += fmt.Sprintf(`%s
  AS%d
  %s
  %s
  
  AS%d
  %s
  %s

`, ixA.Name, a, ixA.Ipaddr4, ixA.Ipaddr6, b, ixB.Ipaddr4, ixB.Ipaddr6)
				} else {
					out += fmt.Sprintf(`  %s %s:
    asn: %d
    neighbors:
      - %s
      - %s

`, networkBInfo.Name, strings.Split(ixA.Name, ":")[0], b, ixB.Ipaddr4, ixB.Ipaddr6)
				}
			}
		}
	}
	return out
}
