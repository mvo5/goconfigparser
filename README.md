Config File Parser (INI style)
==============================

This parser is build as a go equivalent of the Python ConfigParser
module and is aimed for maximum compatibility for both the file format
and the API. This should make it easy to use existing python style
configuration files from go and also ease the porting of existing
python code.

It implements most of RawConfigParser (i.e. no interpolation) at this
point.

Current Limitations:
--------------------
 * no interpolation
 * no defaults
 * no write support
 * not all API is provided