Full-surface Go Wrapper for Haali's Matroska Parser
---

This Go package matroska implements a wrapper for Haali's Matroska Parser.

This was born out of the need for a simple way to get info and, more importantly,
packets and codec private data easily out of Matroska files via standard Go I/O
interfaces. All of the existing packages seemed far too low level (EBML-level)
or just bad.

Ideally I'll slowly port this to native Go at some point. Any day now.


Documentation
---

Please see the [godoc](http://godoc.org/github.com/dwbuiten/matroska) for API
documentation and examples.
