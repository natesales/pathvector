package templating

import (
	"embed"
	"fmt"
	"github.com/natesales/pathvector/internal/config"
	"github.com/natesales/pathvector/internal/util"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"
)

var protocolNames []string

//go:embed templates/*
var embedFS embed.FS

// ConfigWrapper is passed to the peer template
type ConfigWrapper struct {
	Name   string
	Peer   config.Peer
	Global config.Global
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

	"Empty": func(arr *[]string) bool {
		// Is `arr` empty?
		return arr == nil || len(*arr) == 0
	},

	"UnixTimestamp": func() string {
		// Get current UNIX timestamp
		return strconv.Itoa(int(time.Now().Unix()))
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

	"MapDeref": func(m *map[string]string) map[string]string {
		if m != nil {
			return *m
		}
		return map[string]string{}
	},

	// UniqueProtocolName takes a protocol-safe string and address family and returns a unique protocol name
	"UniqueProtocolName": func(s *string, af string) string {
		protoName := fmt.Sprintf("%sv%s", *s, af)
		i := 1
		for {
			if !util.Contains(protocolNames, protoName) {
				protocolNames = append(protocolNames, protoName)
				return protoName
			}
			protoName = fmt.Sprintf("%sv%s_%d", *s, af, i)
			i++
		}
	},
}

// Templates

var PeerTemplate *template.Template
var GlobalTemplate *template.Template
var UITemplate *template.Template
var VRRPTemplate *template.Template

// Load loads the templates from the embedded filesystem
func Load() error {
	var err error

	// Generate peer template
	PeerTemplate, err = template.New("").Funcs(funcMap).ParseFS(embedFS, "templates/peer.tmpl")
	if err != nil {
		return fmt.Errorf("reading peer template: %v", err)
	}

	// Generate global template
	GlobalTemplate, err = template.New("").Funcs(funcMap).ParseFS(embedFS, "templates/global.tmpl")
	if err != nil {
		return fmt.Errorf("reading global template: %v", err)
	}

	// Generate UI template
	UITemplate, err = template.New("").Funcs(funcMap).ParseFS(embedFS, "templates/ui.tmpl")
	if err != nil {
		return fmt.Errorf("reading UI template: %v", err)
	}

	// Generate VRRP template
	VRRPTemplate, err = template.New("").Funcs(funcMap).ParseFS(embedFS, "templates/vrrp.tmpl")
	if err != nil {
		return fmt.Errorf("reading VRRP template: %v", err)
	}

	return nil // nil error
}

// WriteVRRPConfig writes the VRRP config to a keepalived config file
func WriteVRRPConfig(g *config.Global) error {
	if len(g.VRRPInstances) < 1 {
		log.Infof("No VRRP instances are defined, not writing config")
		return nil
	}

	// Create the VRRP config file
	keepalivedFile, err := os.Create(path.Join(g.KeepalivedConfig))
	if err != nil {
		return fmt.Errorf("create peer specific output file: %v", err)
	}

	// Render the template and write to disk
	err = VRRPTemplate.ExecuteTemplate(keepalivedFile, "vrrp.tmpl", g.VRRPInstances)
	if err != nil {
		return fmt.Errorf("execute template: %v", err)
	}

	return nil
}

// WriteUIFile renders and writes the web UI file
func WriteUIFile(g *config.Global) error {
	// Create the UI output file
	log.Debug("Creating UI output file")
	uiFile, err := os.Create(g.WebUIFile)
	if err != nil {
		return fmt.Errorf("create UI output file: %v", err)
	}
	log.Debug("Finished creating UI file")

	// Render the UI template and write to disk
	log.Debug("Writing UI file")
	err = UITemplate.ExecuteTemplate(uiFile, "ui.tmpl", g)
	if err != nil {
		return fmt.Errorf("execute UI template: %v", err)
	}
	log.Debug("Finished writing UI file")

	return nil // nil error
}

// ReformatBirdConfig reformats a BIRD configuration file to look clean and consistent
func ReformatBirdConfig(input string) string {
	formatted := ""
	for _, line := range strings.Split(input, "\n") {
		if strings.HasSuffix(line, "{") || strings.HasSuffix(line, "[") {
			formatted += "\n"
		}

		if !func(input string) bool {
			for _, chr := range []rune(input) {
				if string(chr) != " " {
					return false
				}
			}
			return true
		}(line) {
			formatted += line + "\n"
		}
	}
	return formatted
}
