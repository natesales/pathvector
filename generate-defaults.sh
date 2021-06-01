#!/bin/bash
# This script automates the process of setting default values of complex
# nested YAML types. The idea is that each type needs to implement the
# Unmarshaler interface (https://pkg.go.dev/gopkg.in/yaml.v2#Unmarshaler),
# which turns out to result in a lot of duplicated code.
#
# Instead, this script extracts all the struct definitions from the config.go
# source file and writes a new source file (defaults.go) containing an
# UnmarshalYAML function for each type.
#
# For more information, see this issue with the go-yaml project:
# https://github.com/go-yaml/yaml/issues/165

echo -n "Generating default-aware YAML unmarshalers..."

cat <<EOF >defaults.go
package main

import "github.com/creasty/defaults"
EOF

config_types=$(grep "type .* struct {" config.go | cut -d " " -f 2)
for type in $config_types; do
  cat <<EOF >>defaults.go

func (c *$type) UnmarshalYAML(unmarshal func(interface{}) error) error {
	defaults.Set(c)

	type plain $type
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	return nil
}
EOF
done

echo "done"
