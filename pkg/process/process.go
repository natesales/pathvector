package process

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"

	"github.com/natesales/pathvector/pkg/bird"
	"github.com/natesales/pathvector/pkg/block"
	"github.com/natesales/pathvector/pkg/config"
	"github.com/natesales/pathvector/pkg/embed"
	"github.com/natesales/pathvector/pkg/irr"
	"github.com/natesales/pathvector/pkg/peeringdb"
	"github.com/natesales/pathvector/pkg/plugin"
	"github.com/natesales/pathvector/pkg/templating"
	"github.com/natesales/pathvector/pkg/util"
)

// categorizeCommunity checks if the community is in standard or large form, or an empty string if invalid
func categorizeCommunity(input string) string {
	// Test if it fits the criteria for a standard community
	input = strings.ReplaceAll(input, ",", ":")
	split := strings.Split(input, ":")
	if len(split) == 2 {
		firstPart, err := strconv.Atoi(split[0])
		if err != nil {
			return ""
		}
		secondPart, err := strconv.Atoi(split[1])
		if err != nil {
			return ""
		}

		if firstPart < 0 || firstPart > 65535 {
			return ""
		}
		if secondPart < 0 || secondPart > 65535 {
			return ""
		}
		return "standard"
	} else if len(split) == 3 {
		firstPart, err := strconv.Atoi(split[0])
		if err != nil {
			return ""
		}
		secondPart, err := strconv.Atoi(split[1])
		if err != nil {
			return ""
		}
		thirdPart, err := strconv.Atoi(split[2])
		if err != nil {
			return ""
		}

		if firstPart < 0 || firstPart > 4294967295 {
			return ""
		}
		if secondPart < 0 || secondPart > 4294967295 {
			return ""
		}
		if thirdPart < 0 || thirdPart > 4294967295 {
			return ""
		}
		return "large"
	}

	return ""
}

// sortCommunities sorts a mixed list of standard and large communities into two existing community  lists
func sortCommunities(communities []string) (standard []string, large []string, err error) {
	for _, community := range communities {
		community = strings.ReplaceAll(community, ":", ",")
		switch categorizeCommunity(community) {
		case "standard":
			standard = append(standard, community)
		case "large":
			if large == nil {
				large = []string{}
			}
			large = append(large, community)
		default:
			return nil, nil, errors.New("Invalid import community: " + community)
		}
	}

	return standard, large, nil
}

func sortCommunitiesPtr(communities *[]string) (*[]string, *[]string, error) {
	if communities == nil {
		return &[]string{}, &[]string{}, nil
	}

	standard, large, err := sortCommunities(*communities)
	if err != nil {
		return nil, nil, err
	}

	return &standard, &large, nil
}

func templateReplacements(in string, peer *config.Peer) string {
	v := reflect.ValueOf(peer)
	for v.Kind() == reflect.Ptr { // Dereference pointer types
		v = v.Elem()
	}
	vType := v.Type()
	for i := 0; i < v.NumField(); i++ {
		key := vType.Field(i).Tag.Get("yaml")
		if key != "-" {
			field := v.Field(i)
			key = "<pathvector." + key + ">"
			if !field.IsZero() {
				val := fmt.Sprintf("%v", field.Elem().Interface())
				log.Tracef("Replacing %s with %s\n", key, val)
				in = strings.ReplaceAll(in, key, val)
			}
		}
	}
	return in
}

// Load loads a configuration file from a YAML file
func Load(configBlob []byte) (*config.Config, error) {
	var c config.Config
	c.Init()
	defaults.MustSet(&c)

	if err := util.YAMLUnmarshalStrict(configBlob, &c); err != nil {
		return nil, fmt.Errorf("YAML unmarshal: %s", err)
	}

	validate := validator.New()
	if err := validate.Struct(&c); err != nil {
		return nil, fmt.Errorf("validation: %s", err)
	}

	// Check for invalid templates
	for templateName, templateData := range c.Templates {
		if templateData.Template != nil && *templateData.Template != "" {
			log.Fatalf("Templates must not have a template field set, but %s does", templateName)
		}
	}

	// Set PeeringDB URL
	peeringdb.Endpoint = c.PeeringDBURL
	log.Debugf("Setting PeeringDB endpoint to %s", peeringdb.Endpoint)

	// Set hostname if empty
	if c.Hostname == "" {
		hostname, err := os.Hostname()
		if err != nil {
			log.Fatalf("Hostname is not defined and unable to get system hostname: %s", err)
		}
		c.Hostname = hostname
	}

	if c.Stun {
		c.NoAnnounce = true
		c.NoAccept = true
	}
	if c.NoAnnounce {
		log.Warn("DANGER: no-announce is set, no routes will be announced to any peer")
	}
	if c.NoAccept {
		log.Warn("DANGER: no-accept is set, no routes will be accepted from any peer")
	}

	for peerName, peerData := range c.Peers {
		// Set sanitized peer name
		peerData.ProtocolName = util.Sanitize(peerName)

		// If any peer has NVRS filtering enabled, mark it for querying.
		if peerData.FilterNeverViaRouteServers != nil {
			c.QueryNVRS = true
		}

		if peerData.NeighborIPs == nil || len(*peerData.NeighborIPs) < 1 {
			log.Fatalf("[%s] has no neighbors defined", peerName)
		}

		peerData.BooleanOptions = &[]string{}

		// Assign values from template
		if peerData.Template != nil && *peerData.Template != "" {
			template := c.Templates[*peerData.Template]
			if template == nil {
				log.Fatalf("Template %s not found", *peerData.Template)
			} else {
				templateValue := reflect.ValueOf(*template)
				peerValue := reflect.ValueOf(c.Peers[peerName]).Elem()

				templateValueType := templateValue.Type()
				for i := 0; i < templateValueType.NumField(); i++ {
					fieldName := templateValueType.Field(i).Name
					peerFieldValue := peerValue.FieldByName(fieldName)
					if fieldName != "Template" { // Ignore the template field
						pVal := reflect.Indirect(peerFieldValue)
						peerHasValueConfigured := pVal.IsValid()
						tValue := templateValue.Field(i)
						templateHasValueConfigured := !tValue.IsNil()
						if templateHasValueConfigured && !peerHasValueConfigured {
							// Use the template's value
							peerFieldValue.Set(templateValue.Field(i))
						}

						log.Tracef("[%s] field: %s template's value: %+v kind: %T templateHasValueConfigured: %v", peerName, fieldName, reflect.Indirect(tValue), tValue.Kind().String(), templateHasValueConfigured)
					}
				}
			}
		} // end peer template processor

		// Set default values
		peerValue := reflect.ValueOf(c.Peers[peerName]).Elem()
		templateValueType := peerValue.Type()
		for i := 0; i < templateValueType.NumField(); i++ {
			fieldName := templateValueType.Field(i).Name
			fieldValue := peerValue.FieldByName(fieldName)
			defaultString := templateValueType.Field(i).Tag.Get("default")
			if defaultString == "" {
				log.Fatalf("Code error: field %s has no default value", fieldName)
			} else if defaultString != "-" {
				log.Tracef("[%s] (before defaulting, after templating) field %s value %+v", peerName, fieldName, reflect.Indirect(fieldValue))
				if fieldValue.IsNil() {
					elemToSwitch := templateValueType.Field(i).Type.Elem().Kind()
					switch elemToSwitch {
					case reflect.String:
						log.Tracef("[%s] setting field %s to value %+v", peerName, fieldName, defaultString)
						fieldValue.Set(reflect.ValueOf(&defaultString))
					case reflect.Int:
						defaultValueInt, err := strconv.Atoi(defaultString)
						if err != nil {
							log.Fatalf("Can't convert '%s' to uint", defaultString)
						}
						log.Tracef("[%s] setting field %s to value %+v", peerName, fieldName, defaultValueInt)
						fieldValue.Set(reflect.ValueOf(&defaultValueInt))
					case reflect.Bool:
						var err error // explicit declaration used to avoid scope issues of defaultValue
						defaultBool, err := strconv.ParseBool(defaultString)
						if err != nil {
							log.Fatalf("Can't parse bool %s", defaultString)
						}
						log.Tracef("[%s] setting field %s to value %+v", peerName, fieldName, defaultBool)
						fieldValue.Set(reflect.ValueOf(&defaultBool))
					case reflect.Struct, reflect.Slice:
						// Ignore structs and slices
					default:
						log.Fatalf("Unknown kind %+v for field %s", elemToSwitch, fieldName)
					}
				} else {
					// Add boolean values to the peer's config
					if templateValueType.Field(i).Type.Elem().Kind() == reflect.Bool {
						*peerData.BooleanOptions = append(*peerData.BooleanOptions, templateValueType.Field(i).Tag.Get("yaml"))
					}
				}
			} else {
				log.Tracef("[%s] skipping field %s with ignored default (-)", peerName, fieldName)
			}
		}

		if peerData.PreImportFilter != nil {
			peerData.PreImportFilter = util.Ptr(templateReplacements(*peerData.PreImportFilter, peerData))
		}
		if peerData.PostImportFilter != nil {
			peerData.PostImportFilter = util.Ptr(templateReplacements(*peerData.PostImportFilter, peerData))
		}
		if peerData.PreImportAccept != nil {
			peerData.PreImportAccept = util.Ptr(templateReplacements(*peerData.PreImportAccept, peerData))
		}
		if peerData.PreExport != nil {
			peerData.PreExport = util.Ptr(templateReplacements(*peerData.PreExport, peerData))
		}
		if peerData.PreExportFinal != nil {
			peerData.PreExportFinal = util.Ptr(templateReplacements(*peerData.PreExportFinal, peerData))
		}

		if peerData.DefaultLocalPref != nil && util.Deref(peerData.OptimizeInbound) {
			log.Fatalf("Both DefaultLocalPref and OptimizeInbound set, Pathvector cannot optimize this peer.")
		}

		if peerData.OnlyAnnounce != nil && util.Deref(peerData.AnnounceAll) {
			log.Fatalf("[%s] only-announce and announce-all cannot both be true", peerName)
		}

		// Categorize prefix-communities
		if peerData.PrefixCommunities != nil {
			// Initialize community maps
			if peerData.PrefixStandardCommunities == nil {
				peerData.PrefixStandardCommunities = &map[string][]string{}
			}
			if peerData.PrefixLargeCommunities == nil {
				peerData.PrefixLargeCommunities = &map[string][]string{}
			}

			for prefix, communities := range *peerData.PrefixCommunities {
				for _, community := range communities {
					community = strings.ReplaceAll(community, ":", ",")
					communityType := categorizeCommunity(community)
					if communityType == "standard" {
						if _, ok := (*peerData.PrefixStandardCommunities)[prefix]; !ok {
							(*peerData.PrefixStandardCommunities)[prefix] = []string{}
						}
						(*peerData.PrefixStandardCommunities)[prefix] = append((*peerData.PrefixStandardCommunities)[prefix], community)
					} else if communityType == "large" {
						if _, ok := (*peerData.PrefixLargeCommunities)[prefix]; !ok {
							(*peerData.PrefixLargeCommunities)[prefix] = []string{}
						}
						(*peerData.PrefixLargeCommunities)[prefix] = append((*peerData.PrefixLargeCommunities)[prefix], community)
					} else {
						return nil, errors.New("Invalid prefix community: " + community)
					}
				}
			}
		}

		// Categorize community-prefs
		if peerData.CommunityPrefs != nil {
			// Initialize community maps
			if peerData.StandardCommunityPrefs == nil {
				peerData.StandardCommunityPrefs = &map[string]uint32{}
			}
			if peerData.LargeCommunityPrefs == nil {
				peerData.LargeCommunityPrefs = &map[string]uint32{}
			}

			for community, pref := range *peerData.CommunityPrefs {
				community = strings.ReplaceAll(community, ":", ",")
				communityType := categorizeCommunity(community)
				if communityType == "standard" {
					(*peerData.StandardCommunityPrefs)[community] = pref
				} else if communityType == "large" {
					(*peerData.LargeCommunityPrefs)[community] = pref
				} else {
					return nil, errors.New("Invalid community pref: " + community)
				}
			}
		}

		// Validate RFC 9234 BGP role
		if peerData.Role != nil {
			peerData.Role = util.Ptr(strings.ReplaceAll(*peerData.Role, "-", "_"))
			if *peerData.Role != "provider" && *peerData.Role != "rs_server" && *peerData.Role != "rs_client" && *peerData.Role != "customer" && *peerData.Role != "peer" {
				return nil, fmt.Errorf("[%s] Invalid BGP role: %s (must be one of provider, rs-server, rs-client, customer, peer)", *peerData.Role, peerName)
			}
		}
		requireRoles := peerData.RequireRoles != nil && *peerData.RequireRoles
		if requireRoles && peerData.Role == nil {
			return nil, fmt.Errorf("[%s] require-roles set but no role specified", peerName)
		}

	} // end peer list

	// Parse origin routes by assembling OriginIPv{4,6} lists by address family
	for _, prefix := range c.Prefixes {
		pfx, _, err := net.ParseCIDR(prefix)
		if err != nil {
			return nil, errors.New("Invalid origin prefix: " + prefix)
		}

		if pfx.To4() == nil { // If IPv6
			c.Prefixes6 = append(c.Prefixes6, prefix)
		} else { // If IPv4
			c.Prefixes4 = append(c.Prefixes4, prefix)
		}
	}

	// Initialize static maps
	c.Kernel.Statics4 = map[string]string{}
	c.Kernel.Statics6 = map[string]string{}

	// Categorize communities
	var err error
	c.Kernel.SRDStandardCommunities, c.Kernel.SRDLargeCommunities, err = sortCommunities(c.Kernel.SRDCommunities)
	if err != nil {
		return nil, fmt.Errorf("invalid SRD community: %v", err)
	}
	c.OriginStandardCommunities, c.OriginLargeCommunities, err = sortCommunities(c.OriginCommunities)
	if err != nil {
		return nil, fmt.Errorf("invalid origin community: %v", err)
	}
	c.ImportStandardCommunities, c.ImportLargeCommunities, err = sortCommunities(c.ImportCommunities)
	if err != nil {
		return nil, fmt.Errorf("invalid import community: %v", err)
	}
	c.ExportStandardCommunities, c.ExportLargeCommunities, err = sortCommunities(c.ExportCommunities)
	if err != nil {
		return nil, fmt.Errorf("invalid export community: %v", err)
	}
	c.LocalStandardCommunities, c.LocalLargeCommunities, err = sortCommunities(c.LocalCommunities)
	if err != nil {
		return nil, fmt.Errorf("invalid local community: %v", err)
	}

	// Parse static routes
	for prefix, nexthop := range c.Kernel.Statics {
		// Handle interface suffix
		var rawNexthop string
		if strings.Contains(nexthop, "%") {
			rawNexthop = strings.Split(nexthop, "%")[0]
		} else {
			rawNexthop = nexthop
		}

		pfx, _, err := net.ParseCIDR(prefix)
		if err != nil {
			return nil, errors.New("Invalid static prefix: " + prefix)
		}
		if net.ParseIP(rawNexthop) == nil {
			return nil, errors.New("Invalid static nexthop: " + rawNexthop)
		}

		if pfx.To4() == nil { // If IPv6
			c.Kernel.Statics6[prefix] = nexthop
		} else { // If IPv4
			c.Kernel.Statics4[prefix] = nexthop
		}
	}

	// Parse BFD configs
	for instanceName, bfdInstance := range c.BFDInstances {
		if net.ParseIP(*bfdInstance.Neighbor) == nil {
			return nil, fmt.Errorf("invalid BFD neighbor %s", *bfdInstance.Neighbor)
		}
		bfdInstance.ProtocolName = util.Sanitize(instanceName)
	}

	// Parse VRRP configs
	for _, vrrpInstance := range c.VRRPInstances {
		// Sort VIPs by address family
		for _, vip := range vrrpInstance.VIPs {
			ip, _, err := net.ParseCIDR(vip)
			if err != nil {
				return nil, errors.New("Invalid VIP: " + vip)
			}

			if ip.To4() == nil { // If IPv6
				vrrpInstance.VIPs6 = append(vrrpInstance.VIPs6, vip)
			} else { // If IPv4
				vrrpInstance.VIPs4 = append(vrrpInstance.VIPs4, vip)
			}
		}

		// Validate vrrpInstance
		if vrrpInstance.State == "primary" {
			vrrpInstance.State = "MASTER"
		} else if vrrpInstance.State == "backup" {
			vrrpInstance.State = "BACKUP"
		} else {
			return nil, errors.New("VRRP state must be 'primary' or 'backup', unexpected " + vrrpInstance.State)
		}
	}

	// Parse RTR server
	if c.RTRServer != "" {
		rtrServerParts := strings.Split(c.RTRServer, ":")
		if len(rtrServerParts) != 2 {
			log.Fatalf("Invalid rtr-server '%s' format should be host:port", rtrServerParts)
		}
		c.RTRServerHost = rtrServerParts[0]
		rtrServerPort, err := strconv.Atoi(rtrServerParts[1])
		if err != nil {
			log.Fatalf("Invalid RTR server port %s", rtrServerParts[1])
		}
		c.RTRServerPort = rtrServerPort
	}

	for _, peerData := range c.Peers {
		// Build static prefix filters
		if peerData.Prefixes != nil {
			for _, prefix := range *peerData.Prefixes {
				pfx, _, err := net.ParseCIDR(prefix)
				if err != nil {
					return nil, errors.New("Invalid prefix: " + prefix)
				}

				if pfx.To4() == nil { // If IPv6
					if peerData.PrefixSet6 == nil {
						peerData.PrefixSet6 = &[]string{}
					}
					pfxSet6 := append(*peerData.PrefixSet6, prefix)
					peerData.PrefixSet6 = &pfxSet6
				} else { // If IPv4
					if peerData.PrefixSet4 == nil {
						peerData.PrefixSet4 = &[]string{}
					}
					pfxSet4 := append(*peerData.PrefixSet4, prefix)
					peerData.PrefixSet4 = &pfxSet4
				}
			}
		}

		// Categorize communities
		peerData.ImportStandardCommunities, peerData.ImportLargeCommunities, err = sortCommunitiesPtr(peerData.ImportCommunities)
		if err != nil {
			return nil, fmt.Errorf("invalid import community: %v", err)
		}
		peerData.ExportStandardCommunities, peerData.ExportLargeCommunities, err = sortCommunitiesPtr(peerData.ExportCommunities)
		if err != nil {
			return nil, fmt.Errorf("invalid export community: %v", err)
		}
		peerData.AnnounceStandardCommunities, peerData.AnnounceLargeCommunities, err = sortCommunitiesPtr(peerData.AnnounceCommunities)
		if err != nil {
			return nil, fmt.Errorf("invalid announce community: %v", err)
		}
		peerData.RemoveStandardCommunities, peerData.RemoveLargeCommunities, err = sortCommunitiesPtr(peerData.RemoveCommunities)
		if err != nil {
			return nil, fmt.Errorf("invalid remove community: %v", err)
		}

		// Check for no originated prefixes but announce-originated enabled
		if len(c.Prefixes) < 1 && *peerData.AnnounceOriginated {
			// No locally originated prefixes are defined, so there's nothing to originate
			*peerData.AnnounceOriginated = false
		}
	} // end peer loop

	// Blocklist
	blocklist := block.Combine(c.Blocklist, c.BlocklistURLs, c.BlocklistFiles)
	bASNs, bPrefixes, err := block.Parse(blocklist)
	if err != nil {
		log.Fatal(err)
	}
	c.BlocklistASNs = bASNs
	c.BlocklistPrefixes = bPrefixes
	log.Debugf("Loaded %d ASNs and %d prefixes into global blocklist", len(c.BlocklistASNs), len(c.BlocklistPrefixes))

	// Run plugins
	if err := plugin.ModifyAll(&c); err != nil {
		log.Fatal(err)
	}

	return &c, nil // nil error
}

// peer processes a single peer
func peer(peerName string, peerData *config.Peer, c *config.Config, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Debugf("Processing AS%d %s", *peerData.ASN, peerName)

	// If a PeeringDB query is required
	if *peerData.AutoImportLimits || *peerData.AutoASSet {
		log.Debugf("[%s] has auto-import-limits or auto-as-set, querying PeeringDB", peerName)

		peeringdb.Update(peerData, c.PeeringDBQueryTimeout, c.PeeringDBAPIKey, true)
	} // end peeringdb query enabled

	// Build IRR prefix sets
	if *peerData.FilterIRR {
		if err := irr.Update(peerData, c.IRRServer, c.IRRQueryTimeout, c.BGPQArgs); err != nil {
			log.Fatal(err)
		}
	}
	if *peerData.AutoASSetMembers {
		membersFromIRR, err := irr.ASMembers(*peerData.ASSet, c.IRRServer, c.IRRQueryTimeout, c.BGPQArgs)
		if err != nil {
			log.Fatal(err)
		}
		if peerData.ASSetMembers == nil {
			peerData.ASSetMembers = &membersFromIRR
		} else {
			newASSetMembers := *peerData.ASSetMembers
			newASSetMembers = append(newASSetMembers, membersFromIRR...)
			peerData.ASSetMembers = &newASSetMembers
		}
	}
	if *peerData.FilterASSet && (peerData.ASSetMembers == nil || len(*peerData.ASSetMembers) < 1) {
		log.Fatalf("peer has filter-as-set enabled but no members in it's as-set")
	}

	util.PrintStructInfo(peerName, peerData)

	// Create peer file
	peerFileName := path.Join(c.CacheDirectory, fmt.Sprintf("AS%d_%s.conf", *peerData.ASN, *util.Sanitize(peerName)))
	peerSpecificFile, err := os.Create(peerFileName)
	if err != nil {
		log.Fatalf("Create peer specific output file: %v", err)
	}

	// Render the template and write to buffer
	var b bytes.Buffer
	log.Debugf("[%s] Writing config", peerName)
	if err := templating.Template.ExecuteTemplate(&b, "peer.tmpl", &templating.Wrapper{
		Name:   peerName,
		Peer:   *peerData,
		Config: *c,
	}); err != nil {
		log.Fatalf("Execute template: %v", err)
	}

	// Reformat config and write template to file
	if _, err := peerSpecificFile.Write([]byte(bird.Reformat(b.String()))); err != nil {
		log.Fatalf("Write template to file: %v", err)
	}

	log.Debugf("[%s] Wrote config", peerName)
}

// Run runs the full data generation procedure
func Run(configFilename, lockFile, version string, noConfigure, dryRun, withdraw bool) {
	// Check lockfile
	if lockFile != "" {
		if _, err := os.Stat(lockFile); err == nil {
			log.Fatal("Lockfile exists, exiting")
		} else if os.IsNotExist(err) {
			// If the lockfile doesn't exist, create it
			log.Debug("Lockfile doesn't exist, creating one")
			//nolint:golint,gosec
			if err := os.WriteFile(lockFile, []byte(""), 0644); err != nil {
				log.Fatalf("Writing lockfile: %v", err)
			}
		} else {
			log.Fatalf("Accessing lockfile: %v", err)
		}
	}

	log.Infof("Starting Pathvector %s", version)
	startTime := time.Now()

	// Load the config file from config file
	log.Debugf("Loading config from %s", configFilename)
	configFile, err := os.ReadFile(configFilename)
	if err != nil {
		log.Fatalf("Reading config file: %s", err)
	}
	c, err := Load(configFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("Finished loading config")

	// Run NVRS query
	if c.QueryNVRS {
		var err error
		c.NVRSASNs, err = peeringdb.NeverViaRouteServers(c.PeeringDBQueryTimeout, c.PeeringDBAPIKey)
		if err != nil {
			log.Fatalf("PeeringDB NVRS query: %s", err)
		}
	}

	// Load templates from embedded filesystem
	log.Debug("Loading templates from embedded filesystem")
	err = templating.Load(embed.FS)
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("Finished loading templates")

	// Create cache directory
	log.Debugf("Making cache directory %s", c.CacheDirectory)
	if err := os.MkdirAll(c.CacheDirectory, os.FileMode(0755)); err != nil {
		log.Fatal(err)
	}

	// Create the global output file
	log.Debug("Creating global config")
	globalFile, err := os.Create(path.Join(c.CacheDirectory, "bird.conf"))
	if err != nil {
		log.Fatalf("Create global BIRD output file: %v", err)
	}
	log.Debug("Finished creating global config file")

	// Render the global template and write to buffer
	log.Debug("Writing global config file")
	if err := templating.Template.ExecuteTemplate(globalFile, "global.tmpl", c); err != nil {
		log.Fatalf("Execute global template: %v", err)
	}
	log.Debug("Finished writing global config file")

	// Remove old manual configs
	if err := util.RemoveFileGlob(path.Join(c.CacheDirectory, "manual*.conf")); err != nil {
		log.Fatalf("Removing old manual config files: %v", err)
	}

	// Copying manual configs
	if err := util.CopyFileToGlob(path.Join(c.BIRDDirectory, "manual*.conf"), c.CacheDirectory); err != nil {
		log.Fatalf("Copying manual config files: %v", err)
	}

	// Remove old peer-specific configs
	if err := util.RemoveFileGlob(path.Join(c.CacheDirectory, "AS*.conf")); err != nil {
		log.Fatalf("Removing old peer config files: %v", err)
	}

	// Print global config
	util.PrintStructInfo("pathvector.global", c)

	if withdraw {
		log.Warn("DANGER: withdraw flag is set, withdrawing all routes")
		c.NoAnnounce = true
	}

	// Iterate over peers
	log.Debug("Processing peers")
	wg := new(sync.WaitGroup)
	for peerName, peerData := range c.Peers {
		wg.Add(1)
		go peer(peerName, peerData, c, wg)
	} // end peer loop
	wg.Wait()

	// Run BIRD config validation
	bird.Validate(c.BIRDBinary, c.CacheDirectory)

	// Copy config file
	log.Debug("Copying Pathvector config file to cache directory")
	if err := util.CopyFile(configFilename, path.Join(c.CacheDirectory, "pathvector.yml")); err != nil {
		log.Fatalf("Copying Pathvector config file to cache directory: %v", err)
	}

	if !dryRun {
		// Write protocol name map
		names := templating.ProtocolNames()
		j, err := json.Marshal(names)
		if err != nil {
			log.Fatalf("Marshalling protocol names: %v", err)
		}
		file := path.Join(c.BIRDDirectory, "protocols.json")
		log.Debugf("Writing protocol names to %s", file)
		//nolint:golint,gosec
		if err := os.WriteFile(file, j, 0644); err != nil {
			log.Fatalf("Writing protocol names: %v", err)
		}

		// Write VRRP config
		templating.WriteVRRPConfig(c.VRRPInstances, c.KeepalivedConfig)

		if c.WebUIFile != "" {
			log.Info("Writing web UI")
			templating.WriteUIFile(c)
		}

		bird.MoveCacheAndReconfigure(c.BIRDDirectory, c.CacheDirectory, c.BIRDSocket, noConfigure)
	} // end dry run check

	// Delete lockfile
	if lockFile != "" {
		if err := os.Remove(lockFile); err != nil {
			log.Fatalf("Removing lockfile: %v", err)
		}
	}

	log.Infof("Processed %d sessions over %d peers in %s", countSessions(c.Peers), len(c.Peers), time.Since(startTime).Round(time.Second))
}

func countSessions(peers map[string]*config.Peer) int {
	var count int
	for _, p := range peers {
		count += len(*p.NeighborIPs)
	}
	return count
}
