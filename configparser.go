// A go implementation in the spirit of the python ConfigParser
package goconfigparser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// see python3 configparser.py
var sectionRE = regexp.MustCompile(`^\[(?P<header>[^]]+)\]`)
var optionRE = regexp.MustCompile(`^(?P<option>.*?)\s*(?P<vi>[=|:])\s*(?P<value>.*)$`)
var commentRE = regexp.MustCompile(`^\s*[#;].*$`)

var booleanStates = map[string]bool{
	"1": true, "yes": true, "true": true, "on": true,
	"0": false, "no": false, "false": false, "off": false}

type NoOptionError struct {
	s string
}

func (e NoOptionError) Error() string {
	return e.s
}
func newNoOptionError(section, option string) *NoOptionError {
	return &NoOptionError{s: fmt.Sprintf("No option %s in section %s", option, section)}
}

type NoSectionError struct {
	s string
}

func (e NoSectionError) Error() string {
	return e.s
}
func newNoSectionError(section string) *NoSectionError {
	return &NoSectionError{s: fmt.Sprintf("No section: %s", section)}
}

type ConfigParser struct {
	sections map[string]Section

	// allow data with no ["section"] header via Get("", key)
	AllowNoSectionHeader bool
}

type Section struct {
	options map[string]string
}

// Create a new empty ConfigParser
func New() (cfg *ConfigParser) {
	return &ConfigParser{
		sections: make(map[string]Section)}
}

// Return a string slice of the sections available
func (c *ConfigParser) Sections() (res []string) {
	for k, _ := range c.sections {
		res = append(res, k)
	}
	return res
}

// Return a string slice of the options available in the given section
func (c *ConfigParser) Options(section string) (res []string, err error) {
	sect, ok := c.sections[section]
	if !ok {
		return res, newNoSectionError(section)
	}
	for k, _ := range sect.options {
		res = append(res, k)
	}
	return res, err
}

// Attempt to parse the given string as a configuration
// It may return a error if the parsing fails
func (c *ConfigParser) ReadString(s string) (err error) {
	r := strings.NewReader(s)
	return c.Read(r)
}

// Attempt to parse the given file as a configuration
// It may return a error if the parsing fails
func (c *ConfigParser) ReadFile(path string) (err error) {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	return c.Read(f)
}

// Attempt to parse the given io.Reader as a configuration
// It may return a error if the reading fails
func (c *ConfigParser) Read(r io.Reader) (err error) {
	scanner := bufio.NewScanner(r)

	curSect := ""
	// we allow files with no [section] header in this mode, by
	// default its a error to be aligned with what python configparser
	// is doing
	if c.AllowNoSectionHeader {
		c.sections[""] = Section{
			options: make(map[string]string)}
	}
	for scanner.Scan() {
		line := scanner.Text()
		if commentRE.MatchString(line) {
			// Do nothing
		} else if sectionRE.MatchString(line) {
			matches := sectionRE.FindStringSubmatch(line)
			curSect = matches[1]
			c.sections[curSect] = Section{
				options: make(map[string]string)}
		} else if optionRE.MatchString(line) {
			matches := optionRE.FindStringSubmatch(line)
			key := matches[1]
			value := matches[3]
			value = strings.Replace(value, "%%", "%", -1)
			if _, ok := c.sections[curSect]; !ok {
				return newNoSectionError(curSect)
			}
			c.sections[curSect].options[key] = value
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("ConfigParser scan error %s from %s", err, r)
	}
	return err
}

// Return the option for the given section as string or an error
func (c *ConfigParser) Get(section, option string) (val string, err error) {
	if _, ok := c.sections[section]; !ok {
		return val, newNoSectionError(section)
	}
	sec := c.sections[section]

	if _, ok := sec.options[option]; !ok {
		return val, newNoOptionError(section, option)
	}

	return sec.options[option], err
}

// Return the option for the given section as integer or an error
func (c *ConfigParser) Getint(section, option string) (val int, err error) {
	sv, err := c.Get(section, option)
	if err != nil {
		return val, err
	}
	return strconv.Atoi(sv)
}

// Return the option for the given section as float or an error
func (c *ConfigParser) Getfloat(section, option string) (val float64, err error) {
	sv, err := c.Get(section, option)
	if err != nil {
		return val, err
	}
	return strconv.ParseFloat(sv, 64)
}

// Return the option for the given section as boolean or an error
func (c *ConfigParser) Getbool(section, option string) (val bool, err error) {
	sv, err := c.Get(section, option)
	if err != nil {
		return val, err
	}

	val, ok := booleanStates[strings.ToLower(sv)]
	if !ok {
		return val, errors.New(fmt.Sprintf("No boolean: %s", sv))
	}

	return val, err
}
