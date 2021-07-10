package main

import (
	"embed"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"
)

var protocolNames []string

// wrapper is passed to the peer template
type wrapper struct {
	Name   string
	Peer   Peer
	Config Config
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
			if !contains(protocolNames, protoName) {
				protocolNames = append(protocolNames, protoName)
				return protoName
			}
			protoName = fmt.Sprintf("%sv%s_%d", *s, af, i)
			i++
		}
	},
}

// Templates

var peerTemplate *template.Template
var globalTemplate *template.Template
var uiTemplate *template.Template
var vrrpTemplate *template.Template

// loadTemplates loads the templates from the embedded filesystem
func loadTemplates(fs embed.FS) error {
	var err error

	// Generate peer template
	peerTemplate, err = template.New("").Funcs(funcMap).ParseFS(fs, "templates/peer.tmpl")
	if err != nil {
		return fmt.Errorf("reading peer template: %v", err)
	}

	// Generate global template
	globalTemplate, err = template.New("").Funcs(funcMap).ParseFS(fs, "templates/global.tmpl")
	if err != nil {
		return fmt.Errorf("reading global template: %v", err)
	}

	// Generate UI template
	uiTemplate, err = template.New("").Funcs(funcMap).ParseFS(fs, "templates/ui.tmpl")
	if err != nil {
		return fmt.Errorf("reading UI template: %v", err)
	}

	// Generate VRRP template
	vrrpTemplate, err = template.New("").Funcs(funcMap).ParseFS(fs, "templates/vrrp.tmpl")
	if err != nil {
		return fmt.Errorf("reading VRRP template: %v", err)
	}

	return nil // nil error
}

// writeVRRPConfig writes the VRRP config to a keepalived config file
func writeVRRPConfig(config *Config) {
	if len(config.VRRPInstances) < 1 {
		log.Infof("No VRRP instances are defined, not writing config")
		return
	}

	// Create the VRRP config file
	keepalivedFile, err := os.Create(keepalivedConfig)
	if err != nil {
		log.Fatalf("Create peer specific output file: %v", err)
	}

	// Render the template and write to disk
	err = vrrpTemplate.ExecuteTemplate(keepalivedFile, "vrrp.tmpl", config.VRRPInstances)
	if err != nil {
		log.Fatalf("Execute template: %v", err)
	}
}

// writeUIFile renders and writes the web UI file
func writeUIFile(config *Config) {
	// Create the UI output file
	log.Debug("Creating UI output file")
	uiFileObj, err := os.Create(webUIFile)
	if err != nil {
		log.Fatalf("Create UI output file: %v", err)
	}
	log.Debug("Finished creating UI file")

	// Render the UI template and write to disk
	log.Debug("Writing UI file")
	err = uiTemplate.ExecuteTemplate(uiFileObj, "ui.tmpl", config)
	if err != nil {
		log.Fatalf("Execute UI template: %v", err)
	}
	log.Debug("Finished writing UI file")
}

func reformatBirdConfig(input string) string {
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
