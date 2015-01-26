package goconfigparser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
)

// see python3 configparser.py
var sectionRE = regexp.MustCompile(`\[(?P<header>[^]]+)\]`)
var optionRE = regexp.MustCompile(`^(?P<option>.*?)\s*(?P<vi>[=|:])\s*(?P<value>.*)$`)

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
}

type Section struct {
	options map[string]string
}

func New() (cfg *ConfigParser) {
	return &ConfigParser{
		sections: make(map[string]Section)}
}

func (c *ConfigParser) Sections() (res []string) {
	for k, _ := range c.sections {
		res = append(res, k)
	}
	return res
}

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

func (c *ConfigParser) Read(r io.Reader) (err error) {
	scanner := bufio.NewScanner(r)

	curSect := ""
	for scanner.Scan() {
		line := scanner.Text()
		if sectionRE.MatchString(line) {
			matches := sectionRE.FindStringSubmatch(line)
			curSect = matches[1]
			c.sections[curSect] = Section{
				options: make(map[string]string)}
		} else if optionRE.MatchString(line) {
			matches := optionRE.FindStringSubmatch(line)
			key := matches[1]
			value := matches[3]
			c.sections[curSect].options[key] = value
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("ConfigParser scan error %s from %s", err, r)
	}
	return err
}

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

func (c *ConfigParser) Getint(section, option string) (val int, err error) {
	sv, err := c.Get(section, option)
	if err != nil {
		return val, err
	}
	return strconv.Atoi(sv)
}

func (c *ConfigParser) Getfloat(section, option string) (val float64, err error) {
	sv, err := c.Get(section, option)
	if err != nil {
		return val, err
	}
	return strconv.ParseFloat(sv, 64)
}

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
