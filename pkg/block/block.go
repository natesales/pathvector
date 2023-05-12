package block

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// parseASN parses an ASN into a string and returns -1 if invalid
func parseASN(s string) int {
	s = strings.ToLower(s)
	s = strings.TrimPrefix(s, "as")
	i, err := strconv.Atoi(s)
	if err != nil {
		return -1
	}
	return i
}

// validPrefix checks if a string is a valid IP prefix in CIDR notation
func validPrefix(s string) bool {
	_, _, err := net.ParseCIDR(s)
	return err == nil
}

// validIP returns true if a string is a valid IP address
func validIP(s string) bool {
	return net.ParseIP(s) != nil
}

func removeDuplicate[T string | int | uint32](sliceList []T) []T {
	allKeys := make(map[T]bool)
	var list []T
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

// Parse takes a blocklist and returns a slice of ASNs and prefixes
func Parse(blocklist []string) ([]uint32, []string, error) {
	var asns []uint32
	var prefixes []string

	for _, token := range blocklist {
		// Remove comments
		if strings.HasPrefix(token, "#") {
			continue
		}
		// Remove inline comments
		if strings.Contains(token, "#") {
			token = strings.Split(token, "#")[0]
			token = strings.TrimSpace(token)
		}
		// Skip empty lines
		if token == "" {
			continue
		}
		// Remove whitespace
		token = strings.TrimSpace(token)

		if asn := parseASN(token); asn != -1 {
			log.Debugf("Adding ASN to blocklist: %d", asn)
			asns = append(asns, uint32(asn))
		} else if validPrefix(token) {
			log.Debugf("Adding prefix to blocklist: %s", token)
			prefixes = append(prefixes, token)
		} else if validIP(token) {
			log.Debugf("Adding IP to blocklist: %s", token)

			afiSuffix := "/32"
			if strings.Contains(token, ":") {
				afiSuffix = "/128"
			}
			prefixes = append(prefixes, token+afiSuffix)
		} else {
			return nil, nil, fmt.Errorf("invalid blocklist token: %s", token)
		}
	}

	asns = removeDuplicate(asns)
	prefixes = removeDuplicate(prefixes)

	return asns, prefixes, nil
}

// fromReader reads a blocklist from an io.Reader and returns a slice of strings
func fromReader(reader io.Reader) ([]string, error) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var out []string
	for _, line := range strings.Split(string(body), "\n") {
		if line != "" {
			out = append(out, line)
		}
	}

	return out, nil
}

// Combine aggregates blocklist entries from a manual list, slice of URLs, and slice of files
func Combine(manualBlocklist, blocklistURLs, blocklistFiles []string) []string {
	var out []string

	out = append(out, manualBlocklist...)

	// Fetch blocklist URLs
	for _, url := range blocklistURLs {
		// Fetch the blocklist
		//nolint:gosec
		resp, err := http.Get(url)
		if err != nil {
			log.Warnf("Error fetching blocklist from %s: %s", url, err)
			continue
		}

		entries, err := fromReader(resp.Body)
		if err != nil {
			log.Warnf("Error reading blocklist from %s: %s", url, err)
			continue
		}
		resp.Body.Close()
		out = append(out, entries...)
	}

	// Fetch blocklist files
	for _, file := range blocklistFiles {
		f, err := os.Open(file)
		if err != nil {
			log.Warnf("Error opening blocklist file %s: %s", file, err)
			continue
		}

		entries, err := fromReader(f)
		if err != nil {
			log.Warnf("Error reading blocklist file %s: %s", file, err)
			continue
		}
		f.Close()
		out = append(out, entries...)
	}

	return out
}
