package irr

// This file implements verify-irr-policy (natesales/pathvector#165): checking that a peer's RPSL aut-num object
// still documents a BGP session with us, i.e. it imports routes from us and exports routes to us.
//
// Only the subset of RPSL (RFC 2622) and RPSL-ng (RFC 4012) policy syntax needed to answer that question is parsed:
// the peer expression after each "from"/"to" keyword, whether the line has an "accept"/"announce" clause, and the
// address families from an mp-import/mp-export "afi" list. Filters and actions are not evaluated.

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// PolicyDirection is the direction of an RPSL policy attribute
type PolicyDirection string

const (
	// PolicyImport is an import or mp-import attribute
	PolicyImport PolicyDirection = "import"
	// PolicyExport is an export or mp-export attribute
	PolicyExport PolicyDirection = "export"
)

// PolicyRule is a parsed import, export, mp-import or mp-export attribute
type PolicyRule struct {
	Direction PolicyDirection
	// IPv4 and IPv6 are true if the rule applies to IPv4/IPv6 unicast
	IPv4, IPv6 bool
	// Peers are the (upper case) AS numbers, as-sets or other peer expressions after "from" (import) or "to" (export)
	Peers []string
	// HasAction is true if the rule accepts (import) or announces (export) something other than NOT ANY
	HasAction bool
}

// PolicyResult is whether an aut-num's policy documents a session with an ASN, per direction and address family
type PolicyResult struct {
	Import4, Export4, Import6, Export6 bool
}

var (
	asnRegex = regexp.MustCompile(`^AS\d+$`)
	// Tokens that end an afi list
	afiListEnd = map[string]bool{"from": true, "to": true, "protocol": true, "into": true, "accept": true, "announce": true, "action": true, "{": true}
)

// rpslAttribute is a single attribute of an RPSL object, with continuation lines joined
type rpslAttribute struct {
	name, value string
}

// parseRPSLObjects splits a whois response into RPSL objects (separated by blank lines), joining continuation lines
// (starting with whitespace or "+") and stripping server comments (lines starting with "%") and "#" comments.
func parseRPSLObjects(response string) [][]rpslAttribute {
	var objects [][]rpslAttribute
	var current []rpslAttribute
	flush := func() {
		if len(current) > 0 {
			objects = append(objects, current)
			current = nil
		}
	}

	for _, line := range strings.Split(strings.ReplaceAll(response, "\r\n", "\n"), "\n") {
		if strings.HasPrefix(line, "%") {
			continue
		}
		if strings.TrimSpace(line) == "" {
			flush()
			continue
		}
		if i := strings.Index(line, "#"); i != -1 {
			line = line[:i]
		}

		if line[0] == ' ' || line[0] == '\t' || line[0] == '+' {
			// Continuation of the previous attribute
			if len(current) > 0 {
				current[len(current)-1].value += " " + strings.TrimSpace(strings.TrimPrefix(line, "+"))
			}
			continue
		}
		name, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		current = append(current, rpslAttribute{name: strings.ToLower(strings.TrimSpace(name)), value: strings.TrimSpace(value)})
	}
	flush()
	return objects
}

// tokenize splits an RPSL policy expression into lower case tokens, separating punctuation
func tokenize(s string) []string {
	r := strings.NewReplacer(";", " ; ", ",", " , ", "{", " { ", "}", " } ", "(", " ( ", ")", " ) ")
	return strings.Fields(strings.ToLower(r.Replace(s)))
}

// afiCovers returns which unicast address families an RPSL afi value covers
func afiCovers(afi string) (v4, v6 bool) {
	switch afi {
	case "any", "any.unicast":
		return true, true
	case "ipv4", "ipv4.unicast":
		return true, false
	case "ipv6", "ipv6.unicast":
		return false, true
	}
	return false, false // multicast only or unknown
}

// parsePolicyRule parses the value of an import, export, mp-import or mp-export attribute
func parsePolicyRule(attribute, value string) PolicyRule {
	rule := PolicyRule{Direction: PolicyImport}
	peerKeyword, actionKeyword := "from", "accept"
	if strings.HasSuffix(attribute, "export") {
		rule.Direction = PolicyExport
		peerKeyword, actionKeyword = "to", "announce"
	}

	mp := strings.HasPrefix(attribute, "mp-")
	var afis []string
	tokens := tokenize(value)
	for i := 0; i < len(tokens); i++ {
		switch tokens[i] {
		case "afi":
			// afi <afi>[, <afi>...]
			for i+1 < len(tokens) && !afiListEnd[tokens[i+1]] {
				i++
				if tokens[i] != "," {
					afis = append(afis, tokens[i])
				}
			}
		case peerKeyword:
			// from|to <peer> [OR <peer>...]; anything else after the peer (router addresses, "at") is ignored
			if i+1 < len(tokens) {
				i++
				rule.Peers = append(rule.Peers, strings.ToUpper(tokens[i]))
				for i+2 < len(tokens) && tokens[i+1] == "or" {
					i += 2
					rule.Peers = append(rule.Peers, strings.ToUpper(tokens[i]))
				}
			}
		case actionKeyword:
			// "accept NOT ANY" / "announce NOT ANY" means nothing is exchanged
			if !(i+2 < len(tokens) && tokens[i+1] == "not" && tokens[i+2] == "any") {
				rule.HasAction = true
			}
		}
	}

	switch {
	case len(afis) > 0:
		for _, afi := range afis {
			v4, v6 := afiCovers(afi)
			rule.IPv4 = rule.IPv4 || v4
			rule.IPv6 = rule.IPv6 || v6
		}
	case mp:
		// RFC 4012: an mp-import/mp-export without an afi list applies to all address families
		rule.IPv4, rule.IPv6 = true, true
	default:
		// RFC 2622 import/export only cover IPv4 unicast
		rule.IPv4 = true
	}
	return rule
}

// ParsePolicy parses the import, export, mp-import and mp-export rules of the aut-num objects for asn from a whois
// response. If the response contains more than one aut-num object for the ASN (e.g. from different IRR sources),
// the rules of all of them are returned. found is false if the response has no aut-num object for the ASN.
func ParsePolicy(response string, asn uint32) (rules []PolicyRule, found bool) {
	autNum := fmt.Sprintf("AS%d", asn)
	for _, object := range parseRPSLObjects(response) {
		if object[0].name != "aut-num" || !strings.EqualFold(object[0].value, autNum) {
			continue
		}
		found = true
		for _, attr := range object {
			switch attr.name {
			case "import", "export", "mp-import", "mp-export":
				rules = append(rules, parsePolicyRule(attr.name, attr.value))
			}
		}
	}
	return rules, found
}

// CheckPolicy checks whether policy rules import from and export to localASN, per address family. A rule matches if
// its peer is AS<localASN>, AS-ANY, or an as-set whose members (from expand) include localASN.
// An error is only returned if the result is incomplete because an as-set couldn't be expanded.
func CheckPolicy(rules []PolicyRule, localASN uint32, expand func(asSet string) ([]uint32, error)) (PolicyResult, error) {
	var result PolicyResult
	localAS := fmt.Sprintf("AS%d", localASN)

	set := func(rule PolicyRule) {
		if rule.Direction == PolicyImport {
			result.Import4 = result.Import4 || rule.IPv4
			result.Import6 = result.Import6 || rule.IPv6
		} else {
			result.Export4 = result.Export4 || rule.IPv4
			result.Export6 = result.Export6 || rule.IPv6
		}
	}
	// adds returns true if the rule would add something to the result
	adds := func(rule PolicyRule) bool {
		if rule.Direction == PolicyImport {
			return (rule.IPv4 && !result.Import4) || (rule.IPv6 && !result.Import6)
		}
		return (rule.IPv4 && !result.Export4) || (rule.IPv6 && !result.Export6)
	}

	// First pass: direct matches, which don't need any IRR queries
	var deferred []PolicyRule
	for _, rule := range rules {
		if !rule.HasAction || (!rule.IPv4 && !rule.IPv6) {
			continue
		}
		direct := false
		for _, p := range rule.Peers {
			if p == localAS || p == "AS-ANY" {
				direct = true
			}
		}
		if direct {
			set(rule)
		} else {
			deferred = append(deferred, rule)
		}
	}

	// Second pass: expand as-sets for rules that would still add something
	expanded := map[string]bool{}
	var expandErr error
	for _, rule := range deferred {
		for _, p := range rule.Peers {
			if !adds(rule) {
				break
			}
			if asnRegex.MatchString(p) || !strings.Contains(p, "AS-") {
				continue // another ASN, or a peering-set/unsupported expression
			}
			member, ok := expanded[p]
			if !ok {
				members, err := expand(p)
				if err != nil {
					log.Debugf("Unable to expand as-set %s: %v", p, err)
					expandErr = fmt.Errorf("expanding as-set %s: %w", p, err)
					continue // don't cache, so the outcome stays undetermined
				}
				for _, m := range members {
					if m == localASN {
						member = true
					}
				}
				expanded[p] = member
			}
			if member {
				set(rule)
			}
		}
	}

	// An expansion failure only matters if something is still missing
	if expandErr != nil && result != (PolicyResult{Import4: true, Export4: true, Import6: true, Export6: true}) {
		return result, expandErr
	}
	return result, nil
}

// Missing returns a description of the policy missing from the peer's aut-num for the given address families
func (r PolicyResult) Missing(localASN uint32, v4, v6 bool) []string {
	var missing []string
	check := func(ok bool, format, family string) {
		if !ok {
			missing = append(missing, fmt.Sprintf(format, localASN)+" ("+family+")")
		}
	}
	if v4 {
		check(r.Import4, "import from AS%d accept", "IPv4")
		check(r.Export4, "export to AS%d announce", "IPv4")
	}
	if v6 {
		check(r.Import6, "mp-import afi ipv6 from AS%d accept", "IPv6")
		check(r.Export6, "mp-export afi ipv6 to AS%d announce", "IPv6")
	}
	return missing
}

// QueryAutNum queries an IRR whois server (host or host:port, default port 43) for the aut-num object of an ASN.
// If bgpqArgs restrict the IRR sources with -S, the query is restricted to the same sources.
func QueryAutNum(server string, asn uint32, queryTimeout uint, bgpqArgs string) (string, error) {
	if _, _, err := net.SplitHostPort(server); err != nil {
		server = net.JoinHostPort(server, "43")
	}
	timeout := time.Duration(queryTimeout) * time.Second //nolint:gosec // configured timeout in seconds, far below the int64 range
	conn, err := net.DialTimeout("tcp", server, timeout)
	if err != nil {
		return "", fmt.Errorf("connecting to whois server %s: %w", server, err)
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return "", err
	}

	query := fmt.Sprintf("AS%d", asn)
	if sources := sourcesFlagValue(strings.Fields(bgpqArgs)); sources != "" {
		query = "-s " + sources + " " + query
	}
	log.Debugf("Querying whois server %s for %s", server, query)
	if _, err := fmt.Fprintf(conn, "%s\r\n", query); err != nil {
		return "", fmt.Errorf("whois query to %s: %w", server, err)
	}
	response, err := io.ReadAll(bufio.NewReader(conn))
	if err != nil {
		return "", fmt.Errorf("reading whois response from %s: %w", server, err)
	}
	return string(response), nil
}

// VerifyPolicy checks that the peer's aut-num object imports from and exports to localASN for each address family
// the peer has sessions in (v4, v6). It returns the missing policy (empty if the policy is complete). An error means
// the policy couldn't be checked (whois query failure or as-set expansion failure).
func VerifyPolicy(peerASN, localASN uint32, v4, v6 bool, irrServer string, queryTimeout uint, bgpqArgs string) ([]string, error) {
	response, err := QueryAutNum(irrServer, peerASN, queryTimeout, bgpqArgs)
	if err != nil {
		return nil, err
	}
	rules, found := ParsePolicy(response, peerASN)
	if !found {
		return []string{fmt.Sprintf("aut-num AS%d object (not found)", peerASN)}, nil
	}
	result, err := CheckPolicy(rules, localASN, func(asSet string) ([]uint32, error) {
		return ASMembers(asSet, irrServer, queryTimeout, bgpqArgs)
	})
	missing := result.Missing(localASN, v4, v6)
	if len(missing) > 0 && err != nil {
		// Something is missing, but it might be covered by the as-set that couldn't be expanded
		return nil, err
	}
	return missing, nil
}

// sourcesFlagValue returns the value of the -S (sources) flag in bgpq4 arguments, or an empty string if there is none
func sourcesFlagValue(args []string) string {
	for i, arg := range args {
		if len(arg) < 2 || arg[0] != '-' || arg[1] == '-' {
			continue
		}
		for j, flag := range arg[1:] {
			if flag == 'S' {
				if rest := arg[j+2:]; rest != "" {
					return rest // -SRIPE
				}
				if i+1 < len(args) {
					return args[i+1] // -S RIPE
				}
				return ""
			}
			if strings.ContainsRune(bgpq4ArgFlags, flag) {
				break
			}
		}
	}
	return ""
}
