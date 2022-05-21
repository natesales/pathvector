package process

import (
	"errors"
	"fmt"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/natesales/pathvector/internal/util"
	"github.com/natesales/pathvector/pkg/config"
	"github.com/natesales/pathvector/plugins"
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
	// Set global config defaults
	if err := defaults.Set(&c); err != nil {
		log.Fatal(err)
	}

	if err := yaml.UnmarshalStrict(configBlob, &c); err != nil {
		return nil, errors.New("YAML unmarshal: " + err.Error())
	}

	validate := validator.New()
	if err := validate.Struct(&c); err != nil {
		return nil, errors.New("Validation: " + err.Error())
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

					log.Debugf("[%s] field: %s template's value: %+v kind: %T templateHasValueConfigured: %v", peerName, fieldName, reflect.Indirect(tValue), tValue.Kind().String(), templateHasValueConfigured)
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
	c.Augments.Statics4 = map[string]string{}
	c.Augments.Statics6 = map[string]string{}

	// Categorize communities
	if c.Augments.SRDCommunities != nil {
		for _, community := range c.Augments.SRDCommunities {
			communityType := categorizeCommunity(community)
			if communityType == "standard" {
				if c.Augments.SRDStandardCommunities == nil {
					c.Augments.SRDStandardCommunities = []string{}
				}
				c.Augments.SRDStandardCommunities = append(c.Augments.SRDStandardCommunities, community)
			} else if communityType == "large" {
				if c.Augments.SRDLargeCommunities == nil {
					c.Augments.SRDLargeCommunities = []string{}
				}
				c.Augments.SRDLargeCommunities = append(c.Augments.SRDLargeCommunities, strings.ReplaceAll(community, ":", ","))
			} else {
				return nil, errors.New("Invalid SRD community: " + community)
			}
		}
	}

	// Parse static routes
	for prefix, nexthop := range c.Augments.Statics {
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
			c.Augments.Statics6[prefix] = nexthop
		} else { // If IPv4
			c.Augments.Statics4[prefix] = nexthop
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
	if err := plugins.All(&c); err != nil {
		log.Fatal(err)
	}

	return &c, nil // nil error
}
