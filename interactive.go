package main

import (
	"fmt"
	"github.com/natesales/pathvector/internal/util"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/natesales/pathvector/pkg/config"
)

var (
	verbose = true
	enable  = false
	conf    config.Config
	rline   *readline.Instance
)

type nestedMapContainer struct {
	m map[string]interface{}
}

func completeType(c any, node *nestedMapContainer, target string) {
	if node == nil {
		node = &nestedMapContainer{m: map[string]interface{}{}}
	}
	if node.m == nil {
		node.m = map[string]interface{}{}
	}

	v := reflect.ValueOf(c)
	if v.Kind() == reflect.Ptr { // Dereference pointer types
		v = v.Elem()
	}
	if target != "" {
		completeType(c, &nestedMapContainer{m: node.m[target].(map[string]interface{})}, "")
		return
	}
	vType := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := vType.Field(i)
		key := field.Tag.Get("yaml")
		description := field.Tag.Get("description")
		defaultValue := field.Tag.Get("default")
		if defaultValue == "-" {
			defaultValue = ""
		}

		if description == "" {
			log.Fatalf("%% Code error: %s in %s doesn't have a description: %+v", field.Name, vType.String(), c)
		} else if description != "-" { // Ignore descriptions that are -
			node.m[key] = map[string]interface{}{}
			if strings.Contains(field.Type.String(), "config.") { // If the type is a config struct
				if field.Type.Kind() == reflect.Map {
					newContainer := &nestedMapContainer{m: node.m[key].(map[string]interface{})}
					for _, k := range v.Field(i).MapKeys() {
						log.Debugf("Completing child struct type %s key %s[%s]", field.Type, key, k)
						newContainer.m[k.String()] = map[string]interface{}{}
						completeType(v.Field(i).MapIndex(k).Interface(), newContainer, k.String())
					}
				} else { // If not a map type, insert and recurse
					completeType(v.Field(i).Interface(), &nestedMapContainer{m: node.m[key].(map[string]interface{})}, "")
				}
			}
		}
	}
}

func getConfigValue(c any, item string) (interface{}, error) {
	item = strings.TrimSpace(item)
	log.Debugf("Showing '%s'", item)

	if item == "" { // Global
		return c, nil
	}

	itemSplit := strings.Split(item, " ")
	v := reflect.ValueOf(c)
	if v.Kind() == reflect.Ptr { // Dereference pointer types
		v = v.Elem()
	} else if v.Kind() == reflect.Map {
		for _, k := range v.MapKeys() {
			if k.String() == itemSplit[0] {
				return getConfigValue(v.MapIndex(k).Interface(), strings.Join(itemSplit[1:], " "))
			}
		}
		return nil, fmt.Errorf("%% Configuration item '%s' not found map", item)
	}
	vType := v.Type()
	for i := 0; i < v.NumField(); i++ {
		key := vType.Field(i).Tag.Get("yaml")
		value := v.Field(i).Interface()
		if item == key { // Exact match
			return value, nil
		} else if itemSplit[0] == key {
			return getConfigValue(value, strings.Join(itemSplit[1:], " "))
		}
	}

	return nil, fmt.Errorf("%% Configuration item '%s' not found", item)
}

func setConfigValue(c any, item string) {
	item = strings.TrimSpace(item)
	itemSplit := strings.Split(item, " ")
	targetKey := itemSplit[:len(itemSplit)-1]
	targetValue := itemSplit[len(itemSplit)-1]
	log.Debugf("Attempting to set '%s' to '%s'", targetKey, targetValue)

	v := reflect.ValueOf(c)
	if v.Kind() == reflect.Ptr { // Dereference pointer types
		v = v.Elem()
	}
	vType := v.Type()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		key := vType.Field(i).Tag.Get("yaml")
		value := f.Interface()
		if targetKey[0] == key {
			if len(targetKey) > 1 {
				setConfigValue(value, strings.TrimPrefix(item, targetKey[0]))
			} else { // Exact match
				log.Debugf("Matched. Setting '%s' to '%s'", targetKey, targetValue)
				if f.IsValid() && f.CanSet() {
					switch f.Kind() {
					case reflect.Int, reflect.Int64:
						targetValAsInt, err := strconv.ParseInt(targetValue, 10, 64)
						if err != nil {
							log.Fatalf("%% Unable parse value '%s' as int: %s", targetValue, err)
						}
						if !f.OverflowInt(targetValAsInt) {
							f.SetInt(targetValAsInt)
						}
					case reflect.Uint, reflect.Uint32:
						targetValAsUint, err := strconv.ParseUint(targetValue, 10, 64)
						if err != nil {
							log.Fatalf("%% Unable parse value '%s' as uint: %s", targetValue, err)
						}
						if !f.OverflowUint(targetValAsUint) {
							f.SetUint(targetValAsUint)
						}
					case reflect.String:
						targetValAsString := fmt.Sprintf("%v", targetValue)
						f.SetString(targetValAsString)
					case reflect.Bool:
						if targetValue == "true" {
							f.SetBool(true)
						} else if targetValue == "false" {
							f.SetBool(false)
						} else {
							log.Fatalf("%% Unable parse value '%s' as bool", targetValue)
						}
					default:
						log.Fatalf("%% Unable to set '%s' of type '%s'", item, f.Kind())
					}
				}
			}
		}
	}
}

func printTree(root *nestedMapContainer) {
	fmt.Println("{")
	printTreeRec(root, 1)
	fmt.Println("}")
}

// printTreeRec is the recursive function for printing the tree
func printTreeRec(node *nestedMapContainer, indent int) {
	for k, v := range node.m {
		val := v.(map[string]interface{})

		term := "{},"
		if len(val) > 0 { // has children
			term = "{"
		}

		fmt.Printf("%s\"%s\": %s\n", strings.Repeat("  ", indent), k, term)
		printTreeRec(&nestedMapContainer{m: val}, indent+1)
		if term == "{" {
			fmt.Printf(strings.Repeat("  ", indent) + "},\n")
		}
	}
}

func prettyPrint(a any) {
	o, err := yaml.Marshal(a)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(strings.TrimSpace(string(o)))
}

// completeNode gets a list of prefix completer items for a given node
func completeNode(node *nestedMapContainer) []readline.PrefixCompleterInterface {
	var l []readline.PrefixCompleterInterface
	for k, v := range node.m {
		l = append(l, readline.PcItem(k, completeNode(&nestedMapContainer{m: v.(map[string]interface{})})...))
	}
	return l
}

func prompt(enable bool) string {
	suffix := "> "
	if enable {
		suffix = "# "
	}
	p := "pathvector " + suffix
	hostname, err := os.Hostname()
	if err == nil {
		p = "pathvector (" + hostname + ") " + suffix
	}
	return p
}

func runCommand(line string) {
	line = strings.TrimSpace(line)
	log.Debugf("Processing command '%s'", line)
	switch {
	case line == "enable":
		enable = true
		rline.SetPrompt(prompt(true))
	case line == "disable":
		enable = false
		rline.SetPrompt(prompt(false))
	case line == "show version":
		log.Debugf("Caught command show version")
	case strings.HasPrefix(line, "show"):
		query := strings.TrimPrefix(line, "show")
		item, err := getConfigValue(&conf, query)
		if err != nil {
			fmt.Println(err)
		} else {
			prettyPrint(item)
		}
	case strings.HasPrefix(line, "set"):
		setConfigValue(&conf, strings.TrimPrefix(line, "set"))
	case line == "exit" || line == "quit":
		os.Exit(0)
	case line == "":
	default:
		fmt.Println("% Unknown command: " + line)
	}
}

func main() {
	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	conf.Init()
	conf.Peers["test-peer"] = &config.Peer{ASN: util.IntPtr(3455)}
	conf.Templates["test-template"] = &config.Peer{}
	if err := conf.Default(); err != nil {
		log.Fatal(err)
	}

	var root nestedMapContainer
	completeType(&conf, &root, "")
	printTree(&root)

	topLevelSet := completeNode(&root)
	completer := readline.NewPrefixCompleter(
		readline.PcItem("enable"),
		readline.PcItem("disable"),
		readline.PcItem("show", append(topLevelSet, readline.PcItem("version"))...),
		readline.PcItem("set", topLevelSet...),
		readline.PcItem("delete", topLevelSet...),
		//readline.PcItem("create", topLevelCreate...)
	)

	var err error
	rline, err = readline.NewEx(&readline.Config{
		Prompt:            prompt(enable),
		HistoryFile:       "/tmp/pathvector.cli-history",
		AutoComplete:      completer,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer rline.Close()
	log.SetOutput(rline.Stderr())

	if len(os.Args) > 1 {
		runCommand(strings.Join(os.Args[1:], " "))
	} else {
		for {
			line, err := rline.Readline()
			if err == readline.ErrInterrupt {
				if len(line) == 0 {
					break
				} else {
					continue
				}
			} else if err == io.EOF {
				break
			}

			runCommand(line)
		}
	}
}
