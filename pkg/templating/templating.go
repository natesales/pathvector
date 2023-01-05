package templating

import (
	"embed"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/natesales/pathvector/pkg/config"
	"github.com/natesales/pathvector/pkg/util"
)

var protocolNames []string

// Wrapper is passed to the peer template
type Wrapper struct {
	Name   string
	Peer   config.Peer
	Config config.Config
}

// Template functions
var funcMap = template.FuncMap{
	"Contains": func(s, substr string) bool {
		// String contains
		return strings.Contains(s, substr)
	},

	"Iterate": func(count *int) []int {
		// Create array with `count` entries
		var i int
		var items []int
		for i = 0; i < (*count); i++ {
			items = append(items, i)
		}
		return items
	},

	"BirdSet": func(prefixes []string) string {
		// Build a formatted BIRD prefix list
		output := ""
		for i, prefix := range prefixes {
			output += "  " + prefix
			if i != len(prefixes)-1 {
				output += ",\n"
			}
		}

		return output
	},

	"BirdASSet": func(asns []uint32) string {
		// Build a formatted BIRD AS set
		output := ""
		for i, prefix := range asns {
			output += fmt.Sprintf("  %d", prefix)
			if i != len(asns)-1 {
				output += ",\n"
			}
		}

		return output
	},

	"Empty": func(arr *[]string) bool {
		// Is `arr` empty?
		return arr == nil || len(*arr) == 0
	},

	"Timestamp": func(format string) string {
		// Get current timestamp
		if format == "unix" {
			return strconv.Itoa(int(time.Now().Unix()))
		}
		return time.Now().String()
	},

	"MakeSlice": func(args ...interface{}) []interface{} {
		return args
	},

	"IntCmp": func(i *int, j int) bool {
		return *i == j
	},

	"StringSliceIter": func(slice *[]string) []string {
		if slice != nil {
			return *slice
		}
		return []string{}
	},

	"Uint32SliceDeref": func(slice *[]uint32) []uint32 {
		if slice != nil {
			return *slice
		}
		return []uint32{}
	},

	"StrDeref": func(i *string) string {
		if i != nil {
			return *i
		}
		return ""
	},

	"BoolDeref": func(i *bool) bool {
		if i != nil {
			return *i
		}
		return false
	},

	"IntDeref": func(i *int) int {
		if i != nil {
			return *i
		}
		return 0
	},

	"UintDeref": func(i *uint) uint {
		if i != nil {
			return *i
		}
		return 0
	},

	"MapDeref": func(m *map[string]string) map[string]string {
		if m != nil {
			return *m
		}
		return map[string]string{}
	},

	"Uint32MapDeref": func(m *map[uint32]uint32) map[uint32]uint32 {
		if m != nil {
			return *m
		}
		return map[uint32]uint32{}
	},

	"StrSliceMapDeref": func(m *map[string][]string) map[string][]string {
		if m != nil {
			return *m
		}
		return map[string][]string{}
	},

	"StrSliceDeref": func(s *[]string) []string {
		if s != nil {
			return *s
		}
		return []string{}
	},

	"StrSliceJoin": func(s *[]string) string {
		if s != nil {
			return strings.Join(*s, ", ")
		}
		return ""
	},

	// UniqueProtocolName takes a protocol-safe string and address family and returns a unique protocol name
	"UniqueProtocolName": func(s *string, af string, asn *int) string {
		protoName := fmt.Sprintf("%s_AS%d_v%s", *s, *asn, af)
		i := 1
		for {
			if !util.Contains(protocolNames, protoName) {
				protocolNames = append(protocolNames, protoName)
				return protoName
			}
			protoName = fmt.Sprintf("%s_AS%d_v%s_%d", *s, *asn, af, i)
			i++
		}
	},

	"SplitFirst": func(s string, delim string) string {
		return strings.Split(s, delim)[0]
	},
}

// Templates

var PeerTemplate *template.Template
var GlobalTemplate *template.Template
var UITemplate *template.Template
var VRRPTemplate *template.Template

// Load loads the templates from the embedded filesystem
func Load(fs embed.FS) error {
	var err error

	// Generate peer template
	PeerTemplate, err = template.New("").Funcs(funcMap).ParseFS(fs, "templates/peer.tmpl")
	if err != nil {
		return fmt.Errorf("reading peer template: %v", err)
	}

	// Generate global template
	GlobalTemplate, err = template.New("").Funcs(funcMap).ParseFS(fs, "templates/global.tmpl")
	if err != nil {
		return fmt.Errorf("reading global template: %v", err)
	}

	// Generate UI template
	UITemplate, err = template.New("").Funcs(funcMap).ParseFS(fs, "templates/ui.tmpl")
	if err != nil {
		return fmt.Errorf("reading UI template: %v", err)
	}

	// Generate VRRP template
	VRRPTemplate, err = template.New("").Funcs(funcMap).ParseFS(fs, "templates/vrrp.tmpl")
	if err != nil {
		return fmt.Errorf("reading VRRP template: %v", err)
	}

	return nil // nil error
}

// WriteVRRPConfig writes the VRRP config to a keepalived config file
func WriteVRRPConfig(instances map[string]*config.VRRPInstance, keepalivedConfig string) {
	if len(instances) < 1 {
		log.Debug("No VRRP instances are defined, not writing config")
		return
	}

	// Create the VRRP config file
	keepalivedFile, err := os.Create(keepalivedConfig)
	if err != nil {
		log.Fatalf("Create keepalived output file: %v", err)
	}

	// Render the template and write to disk
	err = VRRPTemplate.ExecuteTemplate(keepalivedFile, "vrrp.tmpl", instances)
	if err != nil {
		log.Fatalf("Execute template: %v", err)
	}
}

// WriteUIFile renders and writes the web UI file
func WriteUIFile(config *config.Config) {
	// Create the UI output file
	log.Debug("Creating UI output file")
	uiFileObj, err := os.Create(config.WebUIFile)
	if err != nil {
		log.Fatalf("Create UI output file: %v", err)
	}
	log.Debug("Finished creating UI file")

	// Render the UI template and write to disk
	log.Debug("Writing UI file")
	err = UITemplate.ExecuteTemplate(uiFileObj, "ui.tmpl", config)
	if err != nil {
		log.Fatalf("Execute UI template: %v", err)
	}
	log.Debug("Finished writing UI file")
}
