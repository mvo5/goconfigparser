[![Build Status][travis-image]][travis-url]
[![GoDoc][godoc-image]][godoc-url]
Config File Parser (INI style)
==============================

This parser is build as a go equivalent of the Python ConfigParser
module and is aimed for maximum compatibility for both the file format
and the API. This should make it easy to use existing python style
configuration files from go and also ease the porting of existing
python code.

Example usage:
```golang
package main

import (
	"fmt"

	"github.com/mvo5/goconfigparser"
)

var cfgExample = `[service]
base: something
`

var cfgExample2 = `[service]
base=something
`

func main() {
	cfg := goconfigparser.New()
	cfg.ReadString(cfgExample)
	val, err := cfg.Get("service", "base")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Got value %q\n", val)
}
```


It implements most of RawConfigParser (i.e. no interpolation) at this
point.

Current Limitations:
--------------------
 * no interpolation
 * no defaults
 * no write support
 * not all API is provided

[travis-image]: https://travis-ci.org/mvo5/goconfigparser.svg?branch=master
[travis-url]: https://travis-ci.org/mvo5/goconfigparser

[godoc-image]: https://godoc.org/github.com/mvo5/goconfigparser?status.svg
[godoc-url]: https://godoc.org/github.com/mvo5/goconfigparser
