package goconfigparser

import (
	"strings"
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner
func Test(t *testing.T) { TestingT(t) }

// partition specific testsuite
type ConfigParserTestSuite struct {
}

var _ = Suite(&ConfigParserTestSuite{})

const SIMPLE_INI = `
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
`

func (s *ConfigParserTestSuite) TestSimple(c *C) {
	cfg := New()
	c.Assert(cfg, NotNil)

	err := cfg.Read(strings.NewReader(SIMPLE_INI))
	c.Assert(err, IsNil)

	val, err := cfg.Get("service", "base")
	c.Assert(err, IsNil)
	c.Assert(val, Equals, "system-image.ubuntu.com")

	val, err = cfg.Get("foo", "bar")
	c.Assert(err, IsNil)
	c.Assert(val, Equals, "baz")

	val, err = cfg.Get("foo", "no-such-option")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "No option no-such-option in section foo")

	val, err = cfg.Get("no-such-section", "no-such-value")
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "No section: no-such-section")
}
