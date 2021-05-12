package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/joomcode/errorx"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/jessevdk/go-flags"
	"github.com/kennygrant/sanitize"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

var version = "dev" // set by the build process

// PeeringDbResponse contains the response from a PeeringDB query
type PeeringDbResponse struct {
	Data []PeeringDbData `json:"data"`
}

// PeeringDbData contains the actual data from PeeringDB response
type PeeringDbData struct {
	Name    string `json:"name"`
	AsSet   string `json:"irr_as_set"`
	MaxPfx4 uint   `json:"info_prefixes4"`
	MaxPfx6 uint   `json:"info_prefixes6"`
}

// Config constants
const (
	DefaultIPv4TableSize = 1000000
	DefaultIPv6TableSize = 150000
)

// Flags
var opts struct {
	ConfigFile       string `short:"c" long:"config" description:"Configuration file in YAML, TOML, or JSON format" default:"/etc/bcg/config.yml"`
	Output           string `short:"o" long:"output" description:"Directory to write output files to" default:"/etc/bird/"`
	Socket           string `short:"s" long:"socket" description:"BIRD control socket" default:"/run/bird/bird.ctl"`
	KeepalivedConfig string `short:"k" long:"keepalived-config" description:"Configuration file for keepalived" default:"/etc/keepalived/keepalived.conf"`
	UiFile           string `short:"u" long:"ui-file" description:"File to store web UI" default:"/tmp/bcg-ui.html"`
	NoUi             bool   `short:"n" long:"no-ui" description:"Don't generate web UI"`
	Verbose          bool   `short:"v" long:"verbose" description:"Show verbose log messages"`
	DryRun           bool   `short:"d" long:"dry-run" description:"Don't modify BIRD config"`
	NoConfigure      bool   `long:"no-configure" description:"Don't configure BIRD"`
	ShowVersion      bool   `long:"version" description:"Show version and exit"`
}

// Embedded filesystem

//go:embed templates/*
var embedFs embed.FS

// contains is a linear search on a string array
func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// Query PeeringDB for an ASN
func getPeeringDbData(asn uint) PeeringDbData {
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

	var peeringDbResponse PeeringDbResponse
	if err := json.Unmarshal(body, &peeringDbResponse); err != nil {
		log.Fatalf("PeeringDB JSON Unmarshal: %v", err)
	}

	if len(peeringDbResponse.Data) < 1 {
		log.Fatalf("Peer %d doesn't have a valid PeeringDB entry. Try import-valid or ask the network to update their account.", asn)
	}

	return peeringDbResponse.Data[0]
}

// Use bgpq4 to generate a prefix filter and return only the filter lines
func getPrefixFilter(asSet string, family uint8, irrdb string) []string {
	// Run bgpq4 for BIRD format with aggregation enabled
	log.Infof("Running bgpq4 -h %s -Ab%d %s", irrdb, family, asSet)
	cmd := exec.Command("bgpq4", "-h", irrdb, "-Ab"+strconv.Itoa(int(family)), asSet)
	stdout, err := cmd.Output()
	if err != nil {
		log.Fatalf("bgpq4 error: %v", err.Error())
	}

	// Remove whitespace and commas from output
	output := strings.ReplaceAll(string(stdout), ",\n    ", "\n")

	// Remove array prefix
	output = strings.ReplaceAll(output, "NN = [\n    ", "")

	// Remove array suffix
	output = strings.ReplaceAll(output, "];", "")

	// Check for empty IRR
	if output == "" {
		log.Warnf("Peer with as-set %s has no IPv%d prefixes. Disabled IPv%d connectivity.", asSet, family, family)
		return []string{}
	}

	// Remove whitespace (in this case there should only be trailing whitespace)
	output = strings.TrimSpace(output)

	// Split output by newline
	return strings.Split(output, "\n")
}

// Normalize a string to be filename-safe
func normalize(input string) string {
	// Remove non-alphanumeric characters
	input = sanitize.Path(input)

	// Make uppercase
	input = strings.ToUpper(input)

	// Replace spaces with underscores
	input = strings.ReplaceAll(input, " ", "_")

	// Replace slashes with dashes
	input = strings.ReplaceAll(input, "/", "-")

	return input
}

// printPeerInfo prints a peer's configuration to the log
func printPeerInfo(peerName string, peerData *Peer) {
	// Fields to exclude from print output
	excludedFields := []string{"PrefixSet4", "PrefixSet6", "Name", "SessionGlobal", "PreImport", "PreExport", "PreImportFinal", "PreExportFinal", "QueryTime"}
	s := reflect.ValueOf(peerData).Elem()
	typeOf := s.Type()
	for i := 0; i < s.NumField(); i++ {
		attrName := typeOf.Field(i).Name
		if !(contains(excludedFields, attrName)) {
			log.Infof("[%s] attribute %s = %v\n", peerName, attrName, s.Field(i).Interface())
		}
	}
}

func main() {
	// Parse cli flags
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		if !strings.Contains(err.Error(), "Usage") {
			log.Fatal(err)
		}
		os.Exit(1)
	}

	// Enable debug logging in development releases
	if //noinspection GoBoolExpressions
	version == "devel" || opts.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	if opts.ShowVersion {
		log.Printf("bcg version %s (https://github.com/natesales/bcg)\n", version)
		os.Exit(0)
	}

	log.Infof("Starting bcg %s", version)

	// Load templates from embedded filesystem
	err = loadTemplates(embedFs)
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("Finished loading templates")

	// Load the config file from configFilename flag
	log.Debugf("Loading config from %s", opts.ConfigFile)
	globalConfig, err := loadConfig(opts.ConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	if !opts.DryRun {
		// Create the global output file
		log.Debug("Creating global config")
		globalFile, err := os.Create(path.Join(opts.Output, "bird.conf"))
		if err != nil {
			log.Fatalf("Create global BIRD output file: %v", err)
		}
		log.Debug("Finished creating global config file")

		// Render the global template and write to disk
		log.Debug("Writing global config file")
		err = globalTemplate.ExecuteTemplate(globalFile, "global.tmpl", globalConfig)
		if err != nil {
			log.Fatalf("Execute global template: %v", err)
		}
		log.Debug("Finished writing global config file")

		// Remove old peer-specific configs
		files, err := filepath.Glob(path.Join(opts.Output, "AS*.conf"))
		if err != nil {
			panic(err)
		}
		for _, f := range files {
			if err := os.Remove(f); err != nil {
				log.Fatalf("Removing old config files: %v", err)
			}
		}
	} else {
		log.Info("Dry run is enabled, skipped writing global config and removing old peer configs")
	}

	// Iterate over peers
	for peerName, peerData := range globalConfig.Peers {
		// Add peer prefix if the first character of peerName is a number
		_peerName := strings.ReplaceAll(normalize(peerName), "-", "_")
		if unicode.IsDigit(rune(_peerName[0])) {
			_peerName = "PEER_" + _peerName
		}

		// Set normalized peer name
		peerData.Name = _peerName

		// Set default query time
		peerData.QueryTime = "[No operations performed]"

		log.Infof("Checking config for %s AS%d", peerName, peerData.Asn)

		// Validate peer type
		if !(peerData.Type == "upstream" || peerData.Type == "peer" || peerData.Type == "downstream" || peerData.Type == "import-valid") {
			log.Fatalf("[%s] type attribute is invalid. Must be upstream, peer, downstream, or import-valid", peerName)
		}

		if !peerData.NoPeeringDB {
			// Only query PeeringDB and IRRDB for peers and downstreams, TODO: This should validate upstreams too
			if peerData.Type == "peer" || peerData.Type == "downstream" {
				peerData.QueryTime = time.Now().Format(time.RFC1123)
				peeringDbData := getPeeringDbData(peerData.Asn)

				if peerData.ImportLimit4 == 0 {
					peerData.ImportLimit4 = peeringDbData.MaxPfx4
					log.Infof("[%s] has no IPv4 import limit configured. Setting to %d from PeeringDB", peerName, peeringDbData.MaxPfx4)
				}

				if peerData.ImportLimit6 == 0 {
					peerData.ImportLimit6 = peeringDbData.MaxPfx6
					log.Infof("[%s] has no IPv6 import limit configured. Setting to %d from PeeringDB", peerName, peeringDbData.MaxPfx6)
				}

				// Only set AS-SET from PeeringDB if it isn't configure manually
				if peerData.AsSet == "" {
					// If the as-set has a space in it, split and pick the first element
					if strings.Contains(peeringDbData.AsSet, " ") {
						peeringDbData.AsSet = strings.Split(peeringDbData.AsSet, " ")[0]
						log.Warnf("[%s] has a space in their PeeringDB as-set field. Selecting first element %s", peerName, peeringDbData.AsSet)
					}

					// Trim IRRDB prefix
					if strings.Contains(peeringDbData.AsSet, "::") {
						peerData.AsSet = strings.Split(peeringDbData.AsSet, "::")[1]
						log.Warnf("[%s] has a IRRDB prefix in their PeeringDB as-set field. Using %s", peerName, peerData.AsSet)
					} else {
						peerData.AsSet = peeringDbData.AsSet
					}

					if peeringDbData.AsSet == "" {
						log.Warnf("[%s] has no as-set in PeeringDB, falling back to their ASN (%d)", peerName, peerData.Asn)
						peerData.AsSet = fmt.Sprintf("AS%d", peerData.Asn)
					} else {
						log.Infof("[%s] has no manual AS-SET defined. Setting to %s from PeeringDB\n", peerName, peeringDbData.AsSet)
					}
				} else {
					log.Infof("[%s] has manual AS-SET: %s", peerName, peerData.AsSet)
				}

				peerData.PrefixSet4 = getPrefixFilter(peerData.AsSet, 4, globalConfig.IrrDb)
				peerData.PrefixSet6 = getPrefixFilter(peerData.AsSet, 6, globalConfig.IrrDb)

				// Update the "latest operation" timestamp
				peerData.QueryTime = time.Now().Format(time.RFC1123)
			} else if peerData.Type == "upstream" || peerData.Type == "import-valid" {
				// Check for a zero prefix import limit
				if peerData.ImportLimit4 == 0 {
					peerData.ImportLimit4 = DefaultIPv4TableSize
					log.Infof("[%s] has no IPv4 import limit configured. Setting to %d", peerName, DefaultIPv4TableSize)
				}

				if peerData.ImportLimit6 == 0 {
					peerData.ImportLimit6 = DefaultIPv6TableSize
					log.Infof("[%s] has no IPv6 import limit configured. Setting to %d", peerName, DefaultIPv6TableSize)
				}
			}
		}

		// If as-set is empty and the peer type requires it
		if peerData.AsSet == "" && (peerData.Type == "peer" || peerData.Type == "downstream") {
			log.Fatalf("[%s] has no AS-SET defined and filtering profile requires it.", peerName)
		}

		// Print peer info
		printPeerInfo(peerName, peerData)

		if !opts.DryRun {
			// Create the peer specific file
			peerSpecificFile, err := os.Create(path.Join(opts.Output, "AS"+strconv.Itoa(int(peerData.Asn))+"_"+normalize(peerName)+".conf"))
			if err != nil {
				log.Fatalf("Create peer specific output file: %v", err)
			}

			// Render the template and write to disk
			log.Infof("[%s] Writing config", peerName)
			err = peerTemplate.ExecuteTemplate(peerSpecificFile, "peer.tmpl", &Wrapper{Peer: *peerData, Config: *globalConfig})
			if err != nil {
				log.Fatalf("Execute template: %v", err)
			}

			log.Infof("[%s] Wrote config", peerName)
		} else {
			log.Infof("Dry run is enabled, skipped writing peer config(s)")
		}
	}

	// Write VRRP config
	if opts.DryRun {
		log.Infof("Dry run is enabled, not writing VRRP config")
	}
	if !opts.DryRun && len(globalConfig.VRRPInstances) > 0 {
		log.Infof("no VRRP instances defined, not writing config")
		// Create the peer specific file
		peerSpecificFile, err := os.Create(path.Join(opts.KeepalivedConfig))
		if err != nil {
			log.Fatalf("Create peer specific output file: %v", err)
		}

		// Render the template and write to disk
		err = vrrpTemplate.ExecuteTemplate(peerSpecificFile, "vrrp.tmpl", globalConfig.VRRPInstances)
		if err != nil {
			log.Fatalf("Execute template: %v", err)
		}
	}

	if !opts.DryRun {
		if !opts.NoUi {
			// Create the ui output file
			log.Debug("Creating global config")
			uiFileObj, err := os.Create(opts.UiFile)
			if err != nil {
				log.Fatalf("Create UI output file: %v", err)
			}
			log.Debug("Finished creating UI file")

			// Render the UI template and write to disk
			log.Debug("Writing ui file")
			err = uiTemplate.ExecuteTemplate(uiFileObj, "ui.tmpl", globalConfig)
			if err != nil {
				log.Fatalf("Execute ui template: %v", err)
			}
			log.Debug("Finished writing ui file")
		}

		if !opts.NoConfigure {
			log.Infoln("Reconfiguring BIRD")
			if err = runBirdCommand("configure", opts.Socket); err != nil {
				log.Fatal(err)
			}
		} else {
			log.Infoln("Option --no-configure is set, NOT reconfiguring bird")
		}

		// Configure interfaces
		for ifaceName, ifaceOpts := range globalConfig.Interfaces {
			if ifaceOpts.Dummy {
				log.Infof("Creating new dummy interface: %s", ifaceName)
				linkAttrs := netlink.NewLinkAttrs()
				linkAttrs.Name = ifaceName
				newIface := &netlink.Dummy{LinkAttrs: linkAttrs}
				if err := netlink.LinkAdd(newIface); err != nil {
					log.Warn(errorx.Decorate(err, "dummy interface create"))
				}
			}

			// Get link by name
			link, err := netlink.LinkByName(ifaceName)
			if err != nil {
				log.Fatal(err)
			}
			log.Debugf("found interface %s index %d", ifaceName, link.Attrs().Index)

			// Set MTU
			if ifaceOpts.Mtu != 0 {
				if err := netlink.LinkSetMTU(link, int(ifaceOpts.Mtu)); err != nil {
					log.Warn(errorx.Decorate(err, "set MTU on "+ifaceName))
				}
			}

			// Add addresses
			for _, addr := range ifaceOpts.Addresses {
				nlAddr, err := netlink.ParseAddr(addr.String())
				if err != nil {
					log.Fatal(err) // This should never happen
				}
				if err := netlink.AddrAdd(link, nlAddr); err != nil {
					log.Warn(errorx.Decorate(err, "add address to "+ifaceName))
				}
			}

			// Add interfaces to xdprtr dataplane
			if ifaceOpts.XDPRTR {
				out, err := exec.Command("xdprtrload", ifaceName).Output()
				if err != nil {
					log.Fatalf("xdprtrload: %v", err)
				}
				log.Infof("xdprtrload: " + string(out))
			}

			// Set interface status
			if ifaceOpts.Down {
				if err := netlink.LinkSetDown(link); err != nil {
					log.Fatal(errorx.Decorate(err, "set link down"))
				}
			} else {
				if err := netlink.LinkSetUp(link); err != nil {
					log.Fatal(errorx.Decorate(err, "set link down"))
				}
			}
		}
	}
}
