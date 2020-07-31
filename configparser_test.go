package goconfigparser

import (
	"io/ioutil"
	"sort"
	"strings"
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner
func Test(t *testing.T) { TestingT(t) }

// partition specific testsuite
type ConfigParserTestSuite struct {
	cfg *ConfigParser
}

var _ = Suite(&ConfigParserTestSuite{})

const SAMPLE_INI = `
# comment: text
  ; indented_comment: text

[service]
base: system-image.ubuntu.com
http_port: 80
https_port: 443
channel: ubuntu-core/devel-proposed
device: generic_amd64
build_number: 246
version_detail: ubuntu=20150121,raw-device=20150121,version=246

[foo]
bar: baz
yesbool: On
nobool: off
float: 3.14
no_interpolation: %%no

[testOptions]
One: 1
Two: 2

[complex]
json_object: {"list":["foo","bar","with\nnewline"]}
json_list: ["foo","bar","with\nnewline"]
`

func (s *ConfigParserTestSuite) SetUpTest(c *C) {
	s.cfg = New()
	c.Assert(s.cfg, NotNil)
	err := s.cfg.ReadString(SAMPLE_INI)
	c.Assert(err, IsNil)
}

func (s *ConfigParserTestSuite) TestSection(c *C) {
	sections := s.cfg.Sections()
	sort.Strings(sections)
	c.Assert(sections, DeepEquals, []string{"complex", "foo", "service", "testOptions"})
}

func (s *ConfigParserTestSuite) TestOptions(c *C) {
	options, err := s.cfg.Options("testOptions")
	c.Assert(err, IsNil)
	sort.Strings(options)
	c.Assert(options, DeepEquals, []string{"One", "Two"})
}

func (s *ConfigParserTestSuite) TestGet(c *C) {
	val, err := s.cfg.Get("service", "base")
	c.Assert(err, IsNil)
	c.Assert(val, Equals, "system-image.ubuntu.com")
}

func (s *ConfigParserTestSuite) TestGetEscapeInterpolation(c *C) {
	val, err := s.cfg.Get("foo", "no_interpolation")
	c.Assert(err, IsNil)
	c.Assert(val, Equals, "%no")
}

func (s *ConfigParserTestSuite) TestGetint(c *C) {
	intval, err := s.cfg.Getint("service", "http_port")
	c.Assert(err, IsNil)
	c.Assert(intval, Equals, 80)
}

func (s *ConfigParserTestSuite) TestGetfloat(c *C) {
	intval, err := s.cfg.Getfloat("foo", "float")
	c.Assert(err, IsNil)
	c.Assert(intval, Equals, 3.14)
}

func (s *ConfigParserTestSuite) TestGetbool(c *C) {
	boolval, err := s.cfg.Getbool("foo", "yesbool")
	c.Assert(err, IsNil)
	c.Assert(boolval, Equals, true)

	boolval, err = s.cfg.Getbool("foo", "nobool")
	c.Assert(err, IsNil)
	c.Assert(boolval, Equals, false)

	boolval, err = s.cfg.Getbool("foo", "bar")
	c.Assert(err.Error(), Equals, "No boolean: baz")
}

func (s *ConfigParserTestSuite) TestErrors(c *C) {
	val, err := s.cfg.Get("foo", "bar")
	c.Assert(err, IsNil)
	c.Assert(val, Equals, "baz")

	val, err = s.cfg.Get("foo", "no-such-option")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "No option no-such-option in section foo")

	val, err = s.cfg.Get("no-such-section", "no-such-value")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "No section: no-such-section")
}

func (s *ConfigParserTestSuite) TestAllowNoSection(c *C) {
	s.cfg = New()
	s.cfg.AllowNoSectionHeader = true
	err := s.cfg.Read(strings.NewReader(`foo=bar`))
	c.Assert(err, IsNil)
	val, err := s.cfg.Get("", "foo")
	c.Assert(val, Equals, "bar")
}

func (s *ConfigParserTestSuite) TestReadFile(c *C) {
	tmp, err := ioutil.TempFile("", "")
	c.Assert(err, IsNil)
	tmp.Write([]byte(SAMPLE_INI))

	s.cfg = New()
	err = s.cfg.ReadFile(tmp.Name())
	c.Assert(err, IsNil)
	val, err := s.cfg.Get("foo", "bar")
	c.Assert(val, Equals, "baz")
}

func (s *ConfigParserTestSuite) TestGetComplex(c *C) {
	val, err := s.cfg.Get("complex", "json_object")
	c.Assert(err, IsNil)
	c.Assert(val, Equals, `{"list":["foo","bar","with\nnewline"]}`)

	val, err = s.cfg.Get("complex", "json_list")
	c.Assert(err, IsNil)
	c.Assert(val, Equals, `["foo","bar","with\nnewline"]`)
}
