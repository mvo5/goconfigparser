package goconfigparser

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"regexp"
)

// see python3 configparser.py
var sectionRE = regexp.MustCompile(`\[(?P<header>[^]]+)\]`)
var optionRE = regexp.MustCompile(`^(?P<option>.*?)\s*(?P<vi>[=|:])\s*(?P<value>.*)$`)

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

type NoOptionError struct {
	s string
}

func (e NoOptionError) Error() string {
	return e.s
}

type NoSectionError struct {
	s string
}

func (e NoSectionError) Error() string {
	return e.s
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
		return val, NoSectionError{s: fmt.Sprintf("No section: %s", section)}
	}
	sec := c.sections[section]

	if _, ok := sec.options[option]; !ok {
		return val, NoOptionError{s: fmt.Sprintf("No option %s in section %s", option, section)}
	}

	return sec.options[option], err
}
