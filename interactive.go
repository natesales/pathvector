package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/chzyer/readline"
	"github.com/creasty/defaults"
	"github.com/natesales/pathvector/pkg/config"
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
				if field.Type.Kind() == reflect.Map || field.Type.Kind() == reflect.Slice { // Extract the element if the type is a map or slice and add to set (reflect.Type to bool map)
					//childTypesSet[field.Type.Elem()] = true
				} else {
					completeType(field.Type, &nestedMapContainer{m: node.m[key].(map[string]interface{})})
				}
			}
			//fmt.Printf("config key %s (%s) default: %s validation: %s", key, description, defaultValue, validation)
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

// completeNode gets a list of prefix completer items for a given node
func completeNode(node *nestedMapContainer) []readline.PrefixCompleterInterface {
	var l []readline.PrefixCompleterInterface
	for k, v := range node.m {
		l = append(l, readline.PcItem(k, completeNode(&nestedMapContainer{m: v.(map[string]interface{})})...))
	}
	return l
}

func main() {
	var c config.Config
	if err := defaults.Set(&c); err != nil {
		log.Fatal(err)
	}

	var root nestedMapContainer
	completeType(reflect.TypeOf(config.Config{}), &root)
	//printTree(&root)

	topLevel := completeNode(&root)
	completer := readline.NewPrefixCompleter(
		readline.PcItem("show", topLevel...),
		readline.PcItem("set", topLevel...),
		readline.PcItem("delete", topLevel...),
	)

	prompt := "pathvector > "
	hostname, err := os.Hostname()
	if err == nil {
		prompt = "pathvector (" + hostname + ") > "
	}
	l, err := readline.NewEx(&readline.Config{
		Prompt:          prompt,
		HistoryFile:     "/tmp/pathvector.cli-history",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold: true,
		FuncFilterInputRune: func(r rune) (rune, bool) { // Block Ctrl+Z
			if r == readline.CharCtrlZ {
				return r, false
			}
			return r, true
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	log.SetOutput(l.Stderr())
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "mode "):
			switch line[5:] {
			case "vi":
				l.SetVimMode(true)
			case "emacs":
				l.SetVimMode(false)
			default:
				println("invalid mode:", line[5:])
			}
		case strings.HasPrefix(line, "set"):
			fmt.Println("Setting!")
		case line == "exit" || line == "quit":
			os.Exit(0)
		case line == "":
		default:
			fmt.Println("% Unknown command: " + line)
		}
	}
}
