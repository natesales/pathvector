// Package yang generates a YANG 1.1 data model of the Pathvector configuration
// from the config structs, and converts RFC 7951 JSON instance data of that
// model into the YAML configuration format understood by Pathvector.
package yang

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unicode"

	"github.com/natesales/pathvector/pkg/config"
)

const (
	// ModuleName is the name of the generated YANG module
	ModuleName = "pathvector"
	// Namespace is the namespace URI of the generated YANG module
	Namespace = "urn:pathvector:config"
	// Prefix is the prefix of the generated YANG module
	Prefix = "pv"

	// listKey is the key leaf of lists generated from map[string]*Struct fields
	listKey = "name"
	// mapKey and mapValue are the leaves of lists generated from other maps
	mapKey   = "key"
	mapValue = "value"
)

// configField is a single user-configurable struct field
type configField struct {
	Name        string // YAML key and YANG node name
	Description string
	Default     string
	Type        reflect.Type // Field type with pointer indirection removed
}

// deref removes pointer indirection from a type
func deref(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// yamlKey returns the YAML key of a struct field (without options)
func yamlKey(f reflect.StructField) string {
	return strings.Split(f.Tag.Get("yaml"), ",")[0]
}

// configFields returns the user-configurable fields of a config struct in declaration order
func configFields(t reflect.Type) []configField {
	t = deref(t)
	var fields []configField
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		key := yamlKey(f)
		description := f.Tag.Get("description")
		if key == "" || key == "-" || description == "-" {
			continue
		}
		def := f.Tag.Get("default")
		if def == "-" {
			def = ""
		}
		fields = append(fields, configField{
			Name:        key,
			Description: description,
			Default:     def,
			Type:        deref(f.Type),
		})
	}
	return fields
}

// isNamedList returns true if t is a map of string to struct (pointer), which is modeled as a list keyed by name
func isNamedList(t reflect.Type) bool {
	return t.Kind() == reflect.Map && t.Key().Kind() == reflect.String && deref(t.Elem()).Kind() == reflect.Struct
}

// groupingName converts a Go type name to a YANG grouping name (e.g. VRRPInstance to vrrp-instance)
func groupingName(t reflect.Type) string {
	name := []rune(deref(t).Name())
	var out strings.Builder
	for i, r := range name {
		if unicode.IsUpper(r) && i > 0 {
			prevLower := unicode.IsLower(name[i-1])
			nextLower := i+1 < len(name) && unicode.IsLower(name[i+1])
			if prevLower || (nextLower && unicode.IsUpper(name[i-1])) {
				out.WriteRune('-')
			}
		}
		out.WriteRune(unicode.ToLower(r))
	}
	return out.String()
}

// quote returns s as a YANG double-quoted string
func quote(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\t", `\t`)
	return `"` + r.Replace(s) + `"`
}

// scalarType returns the YANG built-in type name for a Go scalar type
func scalarType(t reflect.Type) (string, error) {
	switch t.Kind() {
	case reflect.Bool:
		return "boolean", nil
	case reflect.String:
		return "string", nil
	case reflect.Int, reflect.Int64:
		return "int64", nil
	case reflect.Int32:
		return "int32", nil
	case reflect.Int16:
		return "int16", nil
	case reflect.Int8:
		return "int8", nil
	case reflect.Uint, reflect.Uint64:
		return "uint64", nil
	case reflect.Uint32:
		return "uint32", nil
	case reflect.Uint16:
		return "uint16", nil
	case reflect.Uint8:
		return "uint8", nil
	case reflect.Float32, reflect.Float64:
		return "decimal64", nil
	default:
		return "", fmt.Errorf("unsupported type %s", t)
	}
}

// writer is an indenting YANG statement writer
type writer struct {
	b      strings.Builder
	indent int
	err    error
}

func (w *writer) line(format string, args ...interface{}) {
	w.b.WriteString(strings.Repeat("  ", w.indent) + fmt.Sprintf(format, args...) + "\n")
}

func (w *writer) open(format string, args ...interface{}) {
	w.line(format+" {", args...)
	w.indent++
}

func (w *writer) close() {
	w.indent--
	w.line("}")
}

func (w *writer) setErr(err error) {
	if w.err == nil {
		w.err = err
	}
}

// typeStmt writes a type statement for a scalar Go type
func (w *writer) typeStmt(t reflect.Type) {
	yt, err := scalarType(t)
	if err != nil {
		w.setErr(err)
		return
	}
	if yt == "decimal64" {
		w.open("type decimal64")
		w.line("fraction-digits 4;")
		w.close()
		return
	}
	w.line("type %s;", yt)
}

// field writes the schema node (leaf, leaf-list, list or container) for a config field
func (w *writer) field(f configField, groupings map[reflect.Type]bool) {
	t := f.Type
	switch {
	case t.Kind() == reflect.Struct:
		w.open("container %s", f.Name)
		w.line("description %s;", quote(f.Description))
		for _, child := range configFields(t) {
			w.field(child, groupings)
		}
		w.close()
	case isNamedList(t):
		elem := deref(t.Elem())
		for _, child := range configFields(elem) {
			if child.Name == listKey {
				w.setErr(fmt.Errorf("%s has a field named %q which collides with the list key", elem, listKey))
			}
		}
		groupings[elem] = true
		w.open("list %s", f.Name)
		w.line("key %s;", quote(listKey))
		w.line("description %s;", quote(f.Description))
		w.open("leaf %s", listKey)
		w.line("type string;")
		w.line("description %s;", quote("Unique name of this entry"))
		w.close()
		w.line("uses %s;", groupingName(elem))
		w.close()
	case t.Kind() == reflect.Map:
		w.open("list %s", f.Name)
		w.line("key %s;", quote(mapKey))
		w.line("description %s;", quote(f.Description))
		w.open("leaf %s", mapKey)
		w.typeStmt(deref(t.Key()))
		w.line("description %s;", quote("Map key"))
		w.close()
		value := deref(t.Elem())
		if value.Kind() == reflect.Slice {
			w.open("leaf-list %s", mapValue)
			w.typeStmt(deref(value.Elem()))
			w.line("ordered-by user;")
		} else {
			w.open("leaf %s", mapValue)
			w.typeStmt(value)
		}
		w.line("description %s;", quote("Map value"))
		w.close()
		w.close()
	case t.Kind() == reflect.Slice:
		w.open("leaf-list %s", f.Name)
		w.typeStmt(deref(t.Elem()))
		w.line("ordered-by user;")
		w.line("description %s;", quote(f.Description))
		w.close()
	default:
		w.open("leaf %s", f.Name)
		w.typeStmt(t)
		if f.Default != "" {
			w.line("default %s;", quote(f.Default))
		}
		w.line("description %s;", quote(f.Description))
		w.close()
	}
}

// Generate returns a YANG 1.1 module describing the Pathvector configuration model
func Generate() (string, error) {
	return generate(reflect.TypeOf(config.Config{}))
}

func generate(root reflect.Type) (string, error) {
	// Render data nodes first to discover which groupings are needed
	body := &writer{indent: 1}
	groupings := map[reflect.Type]bool{}
	for _, f := range configFields(root) {
		body.field(f, groupings)
	}
	if body.err != nil {
		return "", body.err
	}

	// Render groupings sorted by name
	var groupingTypes []reflect.Type
	for t := range groupings {
		groupingTypes = append(groupingTypes, t)
	}
	sort.Slice(groupingTypes, func(i, j int) bool {
		return groupingName(groupingTypes[i]) < groupingName(groupingTypes[j])
	})
	gw := &writer{indent: 1}
	for _, t := range groupingTypes {
		nested := map[reflect.Type]bool{}
		gw.open("grouping %s", groupingName(t))
		gw.line("description %s;", quote(t.Name()+" configuration"))
		for _, f := range configFields(t) {
			gw.field(f, nested)
		}
		gw.close()
		gw.b.WriteString("\n")
		if len(nested) > 0 {
			gw.setErr(fmt.Errorf("named lists nested inside %s are not supported", t))
		}
	}
	if gw.err != nil {
		return "", gw.err
	}

	w := &writer{}
	w.line("// This file is automatically generated by `pathvector yang`. Do not edit.")
	w.open("module %s", ModuleName)
	w.line("yang-version 1.1;")
	w.line("namespace %s;", quote(Namespace))
	w.line("prefix %s;", Prefix)
	w.b.WriteString("\n")
	w.line("organization %s;", quote("Pathvector"))
	w.line("contact %s;", quote("https://github.com/natesales/pathvector"))
	w.line("description")
	w.line("  %s;", quote("Configuration model of Pathvector, a declarative edge routing platform. "+
		"This module is generated from the Pathvector configuration structs. "+
		"Map sections of the YAML configuration (such as peers and templates) are modeled as lists keyed by name, "+
		"and other maps are modeled as lists of key/value entries."))
	w.b.WriteString("\n")
	w.b.WriteString(gw.b.String())
	w.b.WriteString(body.b.String())
	w.close()
	return w.b.String(), nil
}
