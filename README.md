GO SPDX Parser Library
======================

Parser library for SPDX written in Go (golang). It includes a CLI tool that has
features such as convert, validate and format (pretty-print).

The CLI tool is a good example on how to use the /spdx, /tag and /rdf packages
in other Go programs. All the features available in the CLI tool are available
in the library as well.

Features (planned for first release)
------------------------------------

The following are currently done:
- SPDX 1.2 (the only version supported at the moment)
- parsing RDF formats using [goraptor][goraptor].
- Convert to/from rdf and tag formats
- Validate SPDX documents
- HTML validation output (use the -html flag)
- Auto-detect the input format (file extension or first line guessing)
- Format (pretty-print) SPDX documents (tag format)


Downloading and installing
--------------------------

The easiest way is by using `go get`:

    go get github.com/spdx/spdx-go

This downloads and installs the spdx-go tool and library.

Dependencies
------------

* [libraptor2][raptor] for parsing and serializing RDF
* @deltamobile/goraptor fork of [goraptor][goraptor] by [William Waites][ww]

Building and testing
--------------------

Simple as `go build` and `go test`.

Code
----

The origianl code was available at https://github.com/vladvelici/spdx-go 
and now lives at https://github.com/spdx/tools-go/

[raptor]:http://librdf.org/raptor/
[goraptor]:http://github.com/deltamobile/goraptor
[ww]:https://bitbucket.org/ww/goraptor
