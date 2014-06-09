GO SPDX Parser Library
======================

> **Note:** This is a work in progress. The above features and usage are just what is planned for the initial release.


Parser library for SPDX written in Go (golang). It includes a CLI tool that will have features such as convert, validate and format (pretty-print).

The CLI tool is a good example on how to use the /spdx, /tag and /rdf packages in other Go programs. All the features available in the CLI tool are available in the library as well.

Features (planned for first release)
------------------------------------

- SPDX 1.2 (the only version supported at the moment)
- Convert to/from rdf and tag formats
- Validate rdf and tag formats
- Auto-detect the input format (based on file extension)
- Format (pretty-print) SPDX documents (rdf and tag foramts)

Code
----

The code is available in two repositories:

1. Official repository at http://git.spdx.org/spdx-tools-go.git
2. Mirrored on GitHub at https://github.com/vladvelici/spdx-go
