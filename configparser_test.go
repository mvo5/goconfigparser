package goconfigparser

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
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
	c.Assert(sections, DeepEquals, []string{"foo", "service", "testOptions"})
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

func (s *ConfigParserTestSuite) TestAddSection(c *C) {
	c.Assert(s.cfg.AddSection("new-section"), IsNil)
	c.Assert(s.cfg.AddSection("foo"), ErrorMatches, `Section "foo" already exists`)
	_, ok := s.cfg.sections["new-section"]
	c.Assert(ok, Equals, true)
}

func (s *ConfigParserTestSuite) TestHasSection(c *C) {
	c.Assert(s.cfg.HasSection("foo"), Equals, true)
	c.Assert(s.cfg.HasSection("does-not-exist"), Equals, false)
}

func (s *ConfigParserTestSuite) TestHasOption(c *C) {
	c.Assert(s.cfg.HasOption("foo", "bar"), Equals, true)
	c.Assert(s.cfg.HasOption("foo", "does-not-exist"), Equals, false)
	c.Assert(s.cfg.HasOption("does-not-exist", "bar"), Equals, false)
	c.Assert(s.cfg.HasOption("does-not-exist", "does-not-exist"), Equals, false)
}

func (s *ConfigParserTestSuite) TestHasOptionNoSection(c *C) {
	cfg := New()
	cfg.AllowNoSectionHeader = true
	err := cfg.Read(strings.NewReader("one=1"))
	c.Assert(err, IsNil)
	c.Assert(cfg.HasOption("", "one"), Equals, true)
	c.Assert(cfg.HasOption("", "two"), Equals, false)
	c.Assert(cfg.HasOption("foo", "one"), Equals, false)
	c.Assert(cfg.HasOption("foo", "two"), Equals, false)
}

func (s *ConfigParserTestSuite) TestSet(c *C) {
	c.Assert(s.cfg.Set("foo", "one", "1"), IsNil)
	c.Assert(s.cfg.sections["foo"].options["one"], Equals, "1")
	c.Assert(s.cfg.Set("does-not-exist", "one", "1"), ErrorMatches, "No section: does-not-exist")
}

func (s *ConfigParserTestSuite) TestWrite(c *C) {
	for i, tc := range []struct {
		spaces    bool
		noSection bool
		result    string
	}{
		{false, false, "[foo]\none=1\n\n"},
		{false, true, "one=1\n\n"},
		{true, false, "[foo]\none = 1\n\n"},
		{true, true, "one = 1\n\n"},
	} {
		c.Logf("%d: %v %v", i, tc.spaces, tc.noSection)
		cfg := New()
		sect := ""
		cfg.AllowNoSectionHeader = tc.noSection
		if !tc.noSection {
			sect = "foo"
			c.Assert(cfg.AddSection(sect), IsNil)
		}
		var b bytes.Buffer
		c.Assert(cfg.Set(sect, "one", "1"), IsNil)
		c.Assert(cfg.Write(&b, tc.spaces), IsNil)
		c.Assert(b.String(), Equals, tc.result)
	}
}

func (s *ConfigParserTestSuite) TestWriteFile(c *C) {
	filename := filepath.Join(c.MkDir(), "test.ini")
	cfg := New()
	c.Assert(cfg.AddSection("foo"), IsNil)
	c.Assert(cfg.Set("foo", "one", "1"), IsNil)
	c.Assert(cfg.WriteFile(filename, false, 0644), IsNil)
	data, err := ioutil.ReadFile(filename)
	c.Assert(err, IsNil)
	c.Assert(string(data), Equals, "[foo]\none=1\n\n")
}

func (s *ConfigParserTestSuite) TestRemoveOption(c *C) {
	c.Assert(s.cfg.RemoveOption("foo", "does-not-exist"), ErrorMatches, "No option does-not-exist in section foo")
	c.Assert(s.cfg.RemoveOption("does-not-exist", "bar"), ErrorMatches, "No section: does-not-exist")
	c.Assert(s.cfg.HasOption("foo", "bar"), Equals, true)
	c.Assert(s.cfg.RemoveOption("foo", "bar"), Equals, nil)
	c.Assert(s.cfg.HasOption("foo", "bar"), Equals, false)
}

func (s *ConfigParserTestSuite) TestRemoveSection(c *C) {
	c.Assert(s.cfg.RemoveSection("does-not-exist"), ErrorMatches, "No section: does-not-exist")
	c.Assert(s.cfg.HasSection("foo"), Equals, true)
	c.Assert(s.cfg.RemoveSection("foo"), Equals, nil)
	c.Assert(s.cfg.HasSection("foo"), Equals, false)
}
