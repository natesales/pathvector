package portal

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"

	"github.com/natesales/pathvector/internal/config"
)

// session stores a portal BGP session
type session struct {
	Name       string `json:"name"`
	Router     string `json:"router"`
	ASN        uint32 `json:"asn"`
	LocalIP    string `json:"local-ip"`
	NeighborIP string `json:"neighbor-ip"`
	State      string `json:"state"`
}

// Record records a peer session to the peering portal server
func Record(host string, key string, routerHostname string, peers map[string]*config.Peer) error {
	var sessions []session
	for name, peer := range peers {
		for _, neighborIP := range *peer.NeighborIPs {
			log.Debugf("Adding %s", neighborIP)
			localIP := ""
			if peer.Listen != nil {
				localIP = *peer.Listen
			}
			sessions = append(sessions, session{
				Name:       name,
				Router:     routerHostname,
				ASN:        uint32(*peer.ASN),
				LocalIP:    localIP,
				NeighborIP: neighborIP,
			})
		}
	}

	jsonValue, err := json.Marshal(sessions)
	if err != nil {
		return err
	}
	u, err := url.Parse(host)
	if err != nil {
		return err
	}
	u.Path = "/session"
	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(jsonValue))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", key)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	respText, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		return fmt.Errorf("portal server: %s", respText)
	}

	return nil
}
