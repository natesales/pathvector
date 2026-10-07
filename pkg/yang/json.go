package yang

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/natesales/pathvector/pkg/config"
)

// modulePrefix is the RFC 7951 namespace qualifier of top-level member names
const modulePrefix = ModuleName + ":"

// IsRFC7951 returns true if blob is a JSON object whose members are qualified
// with the pathvector module name, as required by RFC 7951 for top-level nodes
func IsRFC7951(blob []byte) bool {
	trimmed := bytes.TrimSpace(blob)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return false
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &top); err != nil {
		return false
	}
	for k := range top {
		if strings.HasPrefix(k, modulePrefix) {
			return true
		}
	}
	return false
}

// FromRFC7951 converts an RFC 7951 JSON instance of the pathvector YANG module
// into an equivalent Pathvector YAML configuration document.
//
// Lists keyed by name (peers, templates, vrrp, bfd, mrt) are converted to maps
// of name to entry, key/value lists are converted to maps, top-level member
// names have their "pathvector:" qualifier removed, and string encoded
// 64-bit integers and decimals are converted to numbers. Plain JSON documents
// that already use the YAML map structure are passed through unchanged.
func FromRFC7951(blob []byte) ([]byte, error) {
	decoder := json.NewDecoder(bytes.NewReader(blob))
	decoder.UseNumber()
	var doc interface{}
	if err := decoder.Decode(&doc); err != nil {
		return nil, fmt.Errorf("JSON decode: %s", err)
	}
	top, ok := doc.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("top-level JSON value must be an object")
	}
	unqualified := map[string]interface{}{}
	for k, v := range top {
		unqualified[strings.TrimPrefix(k, modulePrefix)] = v
	}

	out, err := convert(unqualified, reflect.TypeOf(config.Config{}), "")
	if err != nil {
		return nil, err
	}
	return yaml.Marshal(out)
}

// convert converts a decoded JSON value v to a YAML-marshalable value matching Go type t
func convert(v interface{}, t reflect.Type, path string) (interface{}, error) {
	t = deref(t)
	if v == nil {
		return nil, nil
	}
	switch t.Kind() {
	case reflect.Struct:
		obj, ok := v.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%s: expected object", path)
		}
		fieldTypes := map[string]reflect.Type{}
		for _, f := range configFields(t) {
			fieldTypes[f.Name] = f.Type
		}
		out := map[string]interface{}{}
		for k, val := range obj {
			ft, found := fieldTypes[k]
			if !found {
				out[k] = val // Unknown fields are rejected by the strict YAML loader
				continue
			}
			c, err := convert(val, ft, path+"/"+k)
			if err != nil {
				return nil, err
			}
			out[k] = c
		}
		return out, nil
	case reflect.Map:
		if obj, ok := v.(map[string]interface{}); ok { // Already in map form
			out := map[interface{}]interface{}{}
			for k, val := range obj {
				key, err := convert(k, t.Key(), path)
				if err != nil {
					return nil, err
				}
				c, err := convert(val, t.Elem(), path+"/"+k)
				if err != nil {
					return nil, err
				}
				out[key] = c
			}
			return out, nil
		}
		entries, ok := v.([]interface{})
		if !ok {
			return nil, fmt.Errorf("%s: expected list", path)
		}
		named := isNamedList(t)
		out := map[interface{}]interface{}{}
		for i, e := range entries {
			entry, ok := e.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("%s[%d]: expected object", path, i)
			}
			if named {
				name, ok := entry[listKey].(string)
				if !ok {
					return nil, fmt.Errorf("%s[%d]: missing string %q key", path, i, listKey)
				}
				rest := map[string]interface{}{}
				for k, val := range entry {
					if k != listKey {
						rest[k] = val
					}
				}
				c, err := convert(rest, t.Elem(), path+"/"+name)
				if err != nil {
					return nil, err
				}
				out[name] = c
			} else {
				rawKey, ok := entry[mapKey]
				if !ok {
					return nil, fmt.Errorf("%s[%d]: missing %q", path, i, mapKey)
				}
				key, err := convert(rawKey, t.Key(), path)
				if err != nil {
					return nil, err
				}
				c, err := convert(entry[mapValue], t.Elem(), fmt.Sprintf("%s/%v", path, key))
				if err != nil {
					return nil, err
				}
				out[key] = c
			}
		}
		return out, nil
	case reflect.Slice:
		items, ok := v.([]interface{})
		if !ok {
			return nil, fmt.Errorf("%s: expected list", path)
		}
		out := make([]interface{}, len(items))
		for i, item := range items {
			c, err := convert(item, t.Elem(), fmt.Sprintf("%s[%d]", path, i))
			if err != nil {
				return nil, err
			}
			out[i] = c
		}
		return out, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(numberString(v), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%s: invalid integer %v", path, v)
		}
		return n, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(numberString(v), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%s: invalid unsigned integer %v", path, v)
		}
		return n, nil
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(numberString(v), 64)
		if err != nil {
			return nil, fmt.Errorf("%s: invalid decimal %v", path, v)
		}
		return n, nil
	default:
		return v, nil
	}
}

// numberString returns the textual representation of a JSON number or string encoded number
func numberString(v interface{}) string {
	switch n := v.(type) {
	case json.Number:
		return n.String()
	case string:
		return n
	default:
		return fmt.Sprint(v)
	}
}
