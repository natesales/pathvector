package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/creasty/defaults"
	"github.com/natesales/pathvector/pkg/config"
)

// Function constructor - constructs new function for listing given directory
func listFiles(path string) func(string) []string {
	return func(line string) []string {
		names := make([]string, 0)
		files, _ := ioutil.ReadDir(path)
		for _, f := range files {
			names = append(names, f.Name())
		}
		return names
	}
}

var completer = readline.NewPrefixCompleter(
	readline.PcItem("mode",
		readline.PcItem("vi"),
		readline.PcItem("emacs"),
	),
	readline.PcItem("login"),
	readline.PcItem("say",
		readline.PcItemDynamic(listFiles("./"),
			readline.PcItem("with",
				readline.PcItem("following"),
				readline.PcItem("items"),
			),
		),
		readline.PcItem("hello"),
		readline.PcItem("bye"),
	),
	readline.PcItem("setprompt"),
	readline.PcItem("setpassword"),
	readline.PcItem("bye"),
	readline.PcItem("help"),
	readline.PcItem("go",
		readline.PcItem("build", readline.PcItem("-o"), readline.PcItem("-v")),
		readline.PcItem("install",
			readline.PcItem("-v"),
			readline.PcItem("-vv"),
			readline.PcItem("-vvv"),
		),
		readline.PcItem("test"),
	),
	readline.PcItem("sleep"),
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

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
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
	completer = readline.NewPrefixCompleter(
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

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	setPasswordCfg := l.GenPasswordConfig()
	setPasswordCfg.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
		l.SetPrompt(fmt.Sprintf("Enter password(%v): ", len(line)))
		l.Refresh()
		return nil, 0, false
	})

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
		case line == "mode":
			if l.IsVimMode() {
				println("current mode: vim")
			} else {
				println("current mode: emacs")
			}
		case line == "login":
			pswd, err := l.ReadPassword("please enter your password: ")
			if err != nil {
				break
			}
			println("you enter:", strconv.Quote(string(pswd)))
		case line == "setpassword":
			pswd, err := l.ReadPasswordWithConfig(setPasswordCfg)
			if err == nil {
				println("you set:", strconv.Quote(string(pswd)))
			}
		case strings.HasPrefix(line, "setprompt"):
			if len(line) <= 10 {
				log.Println("setprompt <prompt>")
				break
			}
			l.SetPrompt(line[10:])
		case strings.HasPrefix(line, "say"):
			line := strings.TrimSpace(line[3:])
			if len(line) == 0 {
				log.Println("say what?")
				break
			}
			go func() {
				for range time.Tick(time.Second) {
					log.Println(line)
				}
			}()
		case line == "exit" || line == "quit":
			os.Exit(0)
		case line == "sleep":
			log.Println("sleep 4 second")
			time.Sleep(4 * time.Second)
		case line == "":
		default:
			fmt.Println("% Unknown command: " + line)
		}
	}
}
