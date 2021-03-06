package templating

import (
	"embed"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/joomcode/errorx"
)

// Template functions
var funcMap = template.FuncMap{
	"Contains": func(s, substr string) bool {
		// String contains
		return strings.Contains(s, substr)
	},

	"Iterate": func(count *uint) []uint {
		// Create array with `count` entries
		var i uint
		var Items []uint
		for i = 0; i < (*count); i++ {
			Items = append(Items, i)
		}
		return Items
	},

	"BirdSet": func(filter []string) string {
		// Build a formatted BIRD prefix list
		output := ""
		for i, prefix := range filter {
			output += "    " + prefix
			if i != len(filter)-1 {
				output += ",\n"
			}
		}

		return output
	},

	"NotEmpty": func(arr []string) bool {
		// Is `arr` empty?
		return len(arr) != 0
	},

	"CheckProtocol": func(v4set []string, v6set []string, family string, peerType string) bool {
		if peerType == "downstream" || peerType == "peer" { // Only match IRR filtered peer types
			if family == "4" {
				return len(v4set) != 0
			} else {
				return len(v6set) != 0
			}
		} else { // If the peer type isn't going to be IRR filtered, ignore it.
			return true
		}
	},

	"CurrentTime": func() string {
		// get current timestamp
		return time.Now().Format(time.RFC1123)
	},

	"UnixTimestamp": func() string {
		// get current timestamp in UNIX format
		return strconv.Itoa(int(time.Now().Unix()))
	},
}

// Templates
var PeerTemplate *template.Template
var GlobalTemplate *template.Template
var UiTemplate *template.Template

// Load loads the templates from the embedded filesystem
func Load(fs embed.FS) error {
	var err error

	// Generate peer template
	PeerTemplate, err = template.New("").Funcs(funcMap).ParseFS(fs, "templates/peer.tmpl")
	if err != nil {
		return errorx.Decorate(err, "Reading peer template")
	}

	// Generate global template
	GlobalTemplate, err = template.New("").Funcs(funcMap).ParseFS(fs, "templates/global.tmpl")
	if err != nil {
		return errorx.Decorate(err, "Reading global template")
	}

	// Generate UI template
	UiTemplate, err = template.New("").Funcs(funcMap).ParseFS(fs, "templates/ui.tmpl")
	if err != nil {
		return errorx.Decorate(err, "Reading ui template")
	}

	return nil // nil error
}
