package main

import (
	"fmt"
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

func completeType(t reflect.Type, node *nestedMapContainer) {
	if node == nil {
		node = &nestedMapContainer{m: map[string]interface{}{}}
	}
	if node.m == nil {
		node.m = map[string]interface{}{}
	}

	if t.Kind() == reflect.Ptr { // Dereference pointer types
		t = t.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		description := field.Tag.Get("description")
		key := field.Tag.Get("yaml")
		//validation := field.Tag.Get("validate")
		defaultValue := field.Tag.Get("default")
		if defaultValue == "-" {
			defaultValue = ""
		}

		if description == "" {
			log.Fatalf("Code error: %s doesn't have a description", field.Name)
		} else if description != "-" { // Ignore descriptions that are -
			node.m[key] = map[string]interface{}{}
			if strings.Contains(field.Type.String(), "config.") { // If the type is a config struct
				if field.Type.Kind() == reflect.Map { // Extract the element if the type is a map or slice
					//log.Infof("Completing child struct type %s key %s", field.Type, key)
					// TODO: Handle this by reading the current map items and setting their completions
					//childTypesSet[field.Type.Elem()] = true
				} else {
					completeType(field.Type, &nestedMapContainer{m: node.m[key].(map[string]interface{})})
				}
			}
			//fmt.Printf("config key %s (%s) default: %s validation: %s", key, description, defaultValue, validation)
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
					default:
						log.Fatalf("%% Unable to set '%s' of type '%s'", item, f.Kind())
					}
				}
			}
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

	if err := conf.Init(); err != nil {
		log.Fatal(err)
	}

	var root nestedMapContainer
	completeType(reflect.TypeOf(config.Config{}), &root)
	//printTree(&root)

	topLevel := completeNode(&root)
	completer := readline.NewPrefixCompleter(
		readline.PcItem("enable"),
		readline.PcItem("disable"),
		readline.PcItem("show", append(topLevel, readline.PcItem("version"))...),
		readline.PcItem("set", topLevel...),
		readline.PcItem("delete", topLevel...),
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
