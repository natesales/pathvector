package util

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"unicode"

	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
)

var alphabet = strings.Split("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", "")

// Contains runs a linear search on a string array
func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// Sanitize limits an input string to only uppercase letters and numbers
func Sanitize(input string) *string {
	output := ""
	for _, chr := range strings.ReplaceAll(strings.ToUpper(input), " ", "_") {
		if Contains(alphabet, string(chr)) || string(chr) == "_" {
			output += string(chr)
		}
	}

	// Add peer prefix if the first character of peerName is a number
	if unicode.IsDigit(rune(output[0])) {
		output = "PEER_" + output
	}

	return &output
}

// MoveFile moves a file from a source to destination
func MoveFile(source, destination string) (err error) {
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()
	fi, err := src.Stat()
	if err != nil {
		return err
	}
	flag := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	perm := fi.Mode() & os.ModePerm
	dst, err := os.OpenFile(destination, flag, perm)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	if err != nil {
		dst.Close()
		os.Remove(destination)
		return err
	}
	err = dst.Close()
	if err != nil {
		return err
	}
	err = src.Close()
	if err != nil {
		return err
	}
	err = os.Remove(source)
	if err != nil {
		return err
	}
	return nil
}

// PrintStructInfo prints a configuration struct values
func PrintStructInfo(label string, instance interface{}) {
	// Fields to exclude from print output
	excludedFields := []string{""}
	s := reflect.ValueOf(instance).Elem()
	typeOf := s.Type()
	for i := 0; i < s.NumField(); i++ {
		attrName := typeOf.Field(i).Name
		if !(Contains(excludedFields, attrName)) {
			v := reflect.Indirect(s.Field(i))
			if v.IsValid() {
				log.Debugf("[%s] field %s = %v\n", label, attrName, v)
			}
		}
	}
}

// PrintTable prints a table of data
func PrintTable(header []string, data [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t") // pad with tabs
	table.SetNoWhiteSpace(true)
	table.AppendBulk(data)
	table.Render()
}

// RemoveFileGlob removes files by glob
func RemoveFileGlob(glob string) error {
	files, err := filepath.Glob(glob)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	return nil
}

// Pointer helpers used to write cleaner tests

// StrDeref returns the value of a pointer to a string
func StrDeref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// StrPtr returns a pointer to a string
func StrPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to an int
func IntPtr(i int) *int {
	return &i
}

// BoolPtr returns a pointer to a boolean
func BoolPtr(b bool) *bool {
	return &b
}
