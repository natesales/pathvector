package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
)

var server = "https://api.pathvector.io"

// CheckLicense checks if the license key is valid
func CheckLicense(license string) {
	if license == "" {
		log.Info("No Pathvector license key found. Contact info@pathvector.io for licensing options.")
		return
	}

	u, err := url.Parse(server)
	if err != nil {
		log.Fatal(err)
	}
	u.Path = "/check"
	q := u.Query()
	u.RawQuery = q.Encode()
	log.Debugf("Connecting to %s", u.String())
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-License-Key", license)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	switch resp.StatusCode {
	case http.StatusOK:
		var lic struct {
			Message string `json:"message"`
			Payload string `json:"payload"`
			Name    string `json:"name"`
			Email   string `json:"email"`
			Expires string `json:"expires"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&lic); err != nil {
			log.Fatal(err)
		}
		log.Infof("Pathvector licensed to %s (%s) [%s] expires %s", lic.Name, lic.Email, lic.Payload, lic.Expires)
	case http.StatusForbidden:
		log.Warnf("Invalid license")
	default:
		log.Warnf("error checking license key: %d", resp.StatusCode)
	}
}

// SendVersionAndLicense sends the license key and version to the Pathvector API
func SendVersionAndLicense(license, version string) {
	u, err := url.Parse(server)
	if err != nil {
		log.Fatal(err)
	}
	u.Path = "/metrics"
	q := u.Query()
	u.RawQuery = q.Encode()
	log.Debugf("Connecting to %s", u.String())
	jsonBytes, err := json.Marshal(map[string]string{
		"license": license,
		"version": version,
	})
	if err != nil {
		log.Fatal(err)
	}
	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(jsonBytes))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
}
