package main

import (
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
)

var alphabet = strings.Split("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", "")

// contains runs a linear search on a string array
func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// sanitize limits an input string to only uppercase letters and numbers
func sanitize(input string) *string {
	output := ""
	for _, chr := range []rune(strings.ReplaceAll(strings.ToUpper(input), " ", "_")) {
		if contains(alphabet, string(chr)) || string(chr) == "_" {
			output += string(chr)
		}
	}

	// Add peer prefix if the first character of peerName is a number
	if unicode.IsDigit(rune(output[0])) {
		output = "PEER_" + output
	}

	return &output
}

// categorizeCommunity checks if the community is in standard or large form, or an empty string if invalid
func categorizeCommunity(input string) string {
	// Test if it fits the criteria for a standard community
	standardSplit := strings.Split(input, ",")
	if len(standardSplit) == 2 {
		firstPart, err := strconv.Atoi(standardSplit[0])
		if err != nil {
			return ""
		}
		secondPart, err := strconv.Atoi(standardSplit[1])
		if err != nil {
			return ""
		}

		if firstPart < 0 || firstPart > 65535 {
			return ""
		}
		if secondPart < 0 || secondPart > 65535 {
			return ""
		}
		return "standard"
	}

	// Test if it fits the criteria for a large community
	largeSplit := strings.Split(input, ":")
	if len(largeSplit) == 3 {
		firstPart, err := strconv.Atoi(largeSplit[0])
		if err != nil {
			return ""
		}
		secondPart, err := strconv.Atoi(largeSplit[1])
		if err != nil {
			return ""
		}
		thirdPart, err := strconv.Atoi(largeSplit[2])
		if err != nil {
			return ""
		}

		if firstPart < 0 || firstPart > 4294967295 {
			return ""
		}
		if secondPart < 0 || secondPart > 4294967295 {
			return ""
		}
		if thirdPart < 0 || thirdPart > 4294967295 {
			return ""
		}
		return "large"
	}

	return ""
}

// moveFile moves a file from a source to destination
func moveFile(source, destination string) (err error) {
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

// printStructInfo prints a configuration struct values
func printStructInfo(label string, instance interface{}) {
	// Fields to exclude from print output
	excludedFields := []string{""}
	s := reflect.ValueOf(instance).Elem()
	typeOf := s.Type()
	for i := 0; i < s.NumField(); i++ {
		attrName := typeOf.Field(i).Name
		if !(contains(excludedFields, attrName)) {
			v := reflect.Indirect(s.Field(i))
			if v.IsValid() {
				log.Debugf("[%s] field %s = %v\n", label, attrName, v)
			}
		}
	}
}

// printTable prints a table of data
func printTable(header []string, data [][]string) {
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

// strDeref returns the value of a pointer to a string
func strDeref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
