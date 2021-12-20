package portal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/natesales/pathvector/internal/bird"
	"github.com/natesales/pathvector/pkg/config"
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
func Record(host string, key string, routerHostname string, peers map[string]*config.Peer, birdSocket string) error {
	// Get protocols
	protocols, err := bird.RunCommand("show protocols", birdSocket)
	if err != nil {
		return err
	}

	var sessions []session
	for name, peer := range peers {
		for _, neighborIP := range *peer.NeighborIPs {
			log.Debugf("Adding %s", neighborIP)
			localIP := ""
			if strings.Contains(neighborIP, ":") { // If IPv6
				if peer.Listen6 != nil {
					localIP = *peer.Listen6
				}
			} else { // If IPv4
				if peer.Listen4 != nil {
					localIP = *peer.Listen4
				}
			}
			// Get session state
			state := "UNKNOWN"
			for _, line := range strings.Split(strings.TrimSuffix(protocols, "\n"), "\n") {
				line = strings.TrimSpace(line)
				if strings.Contains(line, *peer.ProtocolName) {
					line = strings.Split(line, "BGP ")[1]
					line = strings.ReplaceAll(line, "---", "")
					line = strings.Title(line)
					space := regexp.MustCompile(`\s+`)
					state = space.ReplaceAllString(line, " ")
					break
				}
			}
			sessions = append(sessions, session{
				Name:       name,
				Router:     routerHostname,
				ASN:        uint32(*peer.ASN),
				LocalIP:    localIP,
				NeighborIP: neighborIP,
				State:      state,
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
	log.Debugf("Posting %s", jsonValue)
	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(jsonValue))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", key)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	respText, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("portal server: %s", respText)
	}

	return nil
}
