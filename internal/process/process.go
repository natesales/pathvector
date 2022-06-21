package process

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/natesales/pathvector/internal/bird"
	"github.com/natesales/pathvector/internal/embed"
	"github.com/natesales/pathvector/internal/irr"
	"github.com/natesales/pathvector/internal/peeringdb"
	"github.com/natesales/pathvector/internal/portal"
	"github.com/natesales/pathvector/internal/templating"
	"github.com/natesales/pathvector/internal/util"
	"github.com/natesales/pathvector/pkg/config"
	"github.com/natesales/pathvector/pkg/plugin"
)

// categorizeCommunity checks if the community is in standard or large form, or an empty string if invalid
func categorizeCommunity(input string) string {
	// Test if it fits the criteria for a standard community
	standardSplit := strings.Split(input, ",")
	if len(standardSplit) == 2 {
		firstPart, err := strconv.Atoi(standardSplit[0])
		if err != nil {
			return ""
		}
		secondPart, err := strconv.Atoi(standardSplit[1])
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
	}

	// Test if it fits the criteria for a large community
	largeSplit := strings.Split(input, ":")
	if len(largeSplit) == 3 {
		firstPart, err := strconv.Atoi(largeSplit[0])
		if err != nil {
			return ""
		}
		secondPart, err := strconv.Atoi(largeSplit[1])
		if err != nil {
			return ""
		}
		thirdPart, err := strconv.Atoi(largeSplit[2])
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

// Load loads a configuration file from a YAML file
func Load(configBlob []byte) (*config.Config, error) {
	var c config.Config
	c.Init()
	defaults.MustSet(&c)

	if err := yaml.UnmarshalStrict(configBlob, &c); err != nil {
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
			}
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

		// Append snippet files
		if peerData.PreImportFile != nil {
			content, err := os.ReadFile(*peerData.PreImportFile)
			if err != nil {
				log.Fatalf("Unable to read pre-import-file: %s", err)
			}
			*peerData.PreImport += "\n" + string(content)
		}
		if peerData.PreExportFile != nil {
			content, err := os.ReadFile(*peerData.PreExportFile)
			if err != nil {
				log.Fatalf("Unable to read pre-export-file: %s", err)
			}
			*peerData.PreExport += "\n" + string(content)
		}

		if peerData.PreImportFinalFile != nil {
			content, err := os.ReadFile(*peerData.PreImportFinalFile)
			if err != nil {
				log.Fatalf("Unable to read pre-import-final-file: %s", err)
			}
			*peerData.PreImportFinal += "\n" + string(content)
		}
		if peerData.PreExportFinalFile != nil {
			content, err := os.ReadFile(*peerData.PreExportFinalFile)
			if err != nil {
				log.Fatalf("Unable to read pre-export-final-file: %s", err)
			}
			*peerData.PreExportFinal += "\n" + string(content)
		}
		if peerData.DefaultLocalPref != nil && peerData.OptimizeInbound != nil {
			log.Fatalf("Both DefaultLocalPref and OptimizeInbound set, Pathvector cannot optimize this peer.")
		}
	}

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
	if c.Kernel.SRDCommunities != nil {
		for _, community := range c.Kernel.SRDCommunities {
			communityType := categorizeCommunity(community)
			if communityType == "standard" {
				if c.Kernel.SRDStandardCommunities == nil {
					c.Kernel.SRDStandardCommunities = []string{}
				}
				c.Kernel.SRDStandardCommunities = append(c.Kernel.SRDStandardCommunities, community)
			} else if communityType == "large" {
				if c.Kernel.SRDLargeCommunities == nil {
					c.Kernel.SRDLargeCommunities = []string{}
				}
				c.Kernel.SRDLargeCommunities = append(c.Kernel.SRDLargeCommunities, strings.ReplaceAll(community, ":", ","))
			} else {
				return nil, errors.New("Invalid SRD community: " + community)
			}
		}
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
		if peerData.ImportCommunities != nil {
			for _, community := range *peerData.ImportCommunities {
				communityType := categorizeCommunity(community)
				if communityType == "standard" {
					if peerData.ImportStandardCommunities == nil {
						peerData.ImportStandardCommunities = &[]string{}
					}
					*peerData.ImportStandardCommunities = append(*peerData.ImportStandardCommunities, community)
				} else if communityType == "large" {
					if peerData.ImportLargeCommunities == nil {
						peerData.ImportLargeCommunities = &[]string{}
					}
					*peerData.ImportLargeCommunities = append(*peerData.ImportLargeCommunities, strings.ReplaceAll(community, ":", ","))
				} else {
					return nil, errors.New("Invalid import community: " + community)
				}
			}
		}

		if peerData.ExportCommunities != nil {
			for _, community := range *peerData.ExportCommunities {
				communityType := categorizeCommunity(community)
				if communityType == "standard" {
					if peerData.ExportStandardCommunities == nil {
						peerData.ExportStandardCommunities = &[]string{}
					}
					*peerData.ExportStandardCommunities = append(*peerData.ExportStandardCommunities, community)
				} else if communityType == "large" {
					if peerData.ExportLargeCommunities == nil {
						peerData.ExportLargeCommunities = &[]string{}
					}
					*peerData.ExportLargeCommunities = append(*peerData.ExportLargeCommunities, strings.ReplaceAll(community, ":", ","))
				} else {
					return nil, errors.New("Invalid export community: " + community)
				}
			}
		}
		if peerData.AnnounceCommunities != nil {
			for _, community := range *peerData.AnnounceCommunities {
				communityType := categorizeCommunity(community)

				if communityType == "standard" {
					if peerData.AnnounceStandardCommunities == nil {
						peerData.AnnounceStandardCommunities = &[]string{}
					}
					*peerData.AnnounceStandardCommunities = append(*peerData.AnnounceStandardCommunities, community)
				} else if communityType == "large" {
					if peerData.AnnounceLargeCommunities == nil {
						peerData.AnnounceLargeCommunities = &[]string{}
					}
					*peerData.AnnounceLargeCommunities = append(*peerData.AnnounceLargeCommunities, strings.ReplaceAll(community, ":", ","))
				} else {
					return nil, errors.New("Invalid announce community: " + community)
				}
			}
		}
		if peerData.RemoveCommunities != nil {
			for _, community := range *peerData.RemoveCommunities {
				communityType := categorizeCommunity(community)

				if communityType == "standard" {
					if peerData.RemoveStandardCommunities == nil {
						peerData.RemoveStandardCommunities = &[]string{}
					}
					*peerData.RemoveStandardCommunities = append(*peerData.RemoveStandardCommunities, community)
				} else if communityType == "large" {
					if peerData.RemoveLargeCommunities == nil {
						peerData.RemoveLargeCommunities = &[]string{}
					}
					*peerData.RemoveLargeCommunities = append(*peerData.RemoveLargeCommunities, strings.ReplaceAll(community, ":", ","))
				} else {
					return nil, errors.New("Invalid remove community: " + community)
				}
			}
		}

		// Check for no originated prefixes but announce-originated enabled
		if len(c.Prefixes) < 1 && *peerData.AnnounceOriginated {
			// No locally originated prefixes are defined, so there's nothing to originate
			*peerData.AnnounceOriginated = false
		}
	} // end peer loop

	// Run plugins
	if err := plugin.ModifyAll(&c); err != nil {
		log.Fatal(err)
	}

	return &c, nil // nil error
}

// peer processes a single peer
func peer(peerName string, peerData *config.Peer, c *config.Config, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("Processing AS%d %s", *peerData.ASN, peerName)

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
			for _, asn := range membersFromIRR {
				newASSetMembers = append(newASSetMembers, asn)
			}
			peerData.ASSetMembers = &newASSetMembers
		}
	}
	if *peerData.FilterASSet && len(*peerData.ASSetMembers) < 1 {
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
	err = templating.PeerTemplate.ExecuteTemplate(&b, "peer.tmpl", &templating.Wrapper{Name: peerName, Peer: *peerData, Config: *c})
	if err != nil {
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
			log.Debugln("Lockfile doesn't exist, creating one")
			//nolint:golint,gosec
			if err := os.WriteFile(lockFile, []byte(""), 0755); err != nil {
				log.Fatalf("Writing lockfile: %v", err)
			}
		} else {
			log.Fatalf("Accessing lockfile: %v", err)
		}
	}

	log.Debugf("Starting pathvector %s", version)

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
	log.Debugln("Finished loading config")

	// Run NVRS query
	if c.QueryNVRS {
		var err error
		c.NVRSASNs, err = peeringdb.NeverViaRouteServers(c.PeeringDBQueryTimeout, c.PeeringDBAPIKey)
		if err != nil {
			log.Fatalf("PeeringDB NVRS query: %s", err)
		}
	}

	// Load templates from embedded filesystem
	log.Debugln("Loading templates from embedded filesystem")
	err = templating.Load(embed.FS)
	if err != nil {
		log.Fatal(err)
	}
	log.Debugln("Finished loading templates")

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
	err = templating.GlobalTemplate.ExecuteTemplate(globalFile, "global.tmpl", c)
	if err != nil {
		log.Fatalf("Execute global template: %v", err)
	}
	log.Debug("Finished writing global config file")

	// Remove old peer-specific configs
	if err := util.RemoveFileGlob(path.Join(c.CacheDirectory, "AS*.conf")); err != nil {
		log.Fatalf("Removing old config files: %v", err)
	}

	// Print global config
	util.PrintStructInfo("pathvector.global", c)

	if withdraw {
		log.Warn("DANGER: withdraw flag is set, withdrawing all routes")
		c.NoAnnounce = true
	}

	// Iterate over peers
	wg := new(sync.WaitGroup)
	for peerName, peerData := range c.Peers {
		wg.Add(1)
		go peer(peerName, peerData, c, wg)
	} // end peer loop
	wg.Wait()

	// Run BIRD config validation
	bird.Validate(c.BIRDBinary, c.CacheDirectory)

	if !dryRun {
		// Write VRRP config
		templating.WriteVRRPConfig(c.VRRPInstances, c.KeepalivedConfig)

		if c.WebUIFile != "" {
			templating.WriteUIFile(c)
		} else {
			log.Infof("Web UI is not defined, NOT writing UI")
		}

		bird.MoveCacheAndReconfigure(c.BIRDDirectory, c.CacheDirectory, c.BIRDSocket, noConfigure)
	} // end dry run check

	// Update portal
	if c.PortalHost != "" {
		log.Infoln("Updating peering portal")
		if err := portal.Record(c.PortalHost, c.PortalKey, c.Hostname, c.Peers, c.BIRDSocket); err != nil {
			log.Fatal(err)
		}
	}

	// Delete lockfile
	if lockFile != "" {
		if err := os.Remove(lockFile); err != nil {
			log.Fatalf("Removing lockfile: %v", err)
		}
	}
}
