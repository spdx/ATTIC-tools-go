/*
spdx-go is a tool for pretty-printing, converting and validating SPDX files.
Formats supported by this tool are RDF (all RDF syntaxes supported by the raptor
library) and Tag. For a full list of formats, see `-help`.

Basic usage
===========

The basic usage of this tool is the following:

    spdx-go <flags> <input file>

One action flag must be specified. Those are:

		-v						# validation
		-c <format>		# conversion
		-p						# pretty-printing (formatting)
		-help					# print the help message and quit
		-version			# print the tool version and quit

For a list of all flags and usage, see `-help`.

Input and output
----------------

The input, if no input file specified, is read from standard input. and input
format (using flag `-i <format>`) must be specified in this case.

Output is, by default, on standard output. The flag `-o <file>` can be used to
specify an output file.

Flag `-w` can be used to overwrite the input file, useful for pretty-printing:

    spdx-go -p -w example.tag

The flag -w creates a temporary file and tries to change the innode of the input
file to the temporary file and then delete the temporary file. If that fails, it
hard copies the temporary file over the input file.

SPDX File formats
-----------------

With every command in this tool, the flag -f <format> makes the parser to use
<format> as the input format. For a list of all valid formats, see `-help`.

The format "rdf" is a special format. For output, it means "xmlrdf-abbrev"; for
input, it means attempt to guess the RDF syntax in the input file (uses raptor's
"guess" parser).

Pretty-print (format) SPDX file
===============================

Use the `-p` flag to pretty-print a SPDX document. Pretty-printing does not
parse the input document into a `spdx.Document` struct but only tokenizes the
input format and pretty-prints the tokens.

This means that invalid SPDX documents may be printed.

A known limitation is the fact that comments in any RDF syntax are dismissed.
The same limitation does not apply to the Tag format, where comments are printed
and formatted.

Example:

		# the -w flag is optional and overwrites the input file
		spdx-go -p -w example.tag

Convert between formats
=======================

Use the `-c <format>` flag to convert to and from supported SPDX formats.

Example:

		# conver example.tag to example.rdf
    spdx-go -c rdf -o example.rdf example.tag

Validate SPDX file
==================

SPDX files in any formats are parsed to `spdx.Document` and thus validated to
the SPDX Specification. Use the `-v` flag to validate documents. Example:

		spdx-go -v example.tag
		spdx-go -v example.rd

HTML output validation
----------------------

To better visualise the validation errors and warnings, HTML output is supported
by this tool. The `-html` flag used in conjunction with the `-v` flag creates a
HTML file that contains the input file and all the validation errors represented
nicely in the page.

If no output file is specified, a temporary file is created and opened in a
(default) browser window. If there is an output file specified (`-o <file>`),
the HTML is written to that file instead and no browser window opened.

Example:

		spdx-go -v -html example.html

Updating licence list
---------------------

The spdx-go tool assumes that a file named `licence-list.txt` exists. However,
this file is not included in the repository but it is quite simple to generate:

		./update-list.sh

The file only contains one licence ID per line. The script generates the most
up-to-date file from the SPDX Licence List git repository. This git repository
has a submodule in `spdx/licence-list` which points to the official SPDX Licence
List repository. For how it works, see the documentation for the `spdx` package.

spdx-tools-go
=============

This tool also serves as an example usage of the SPDX Go Parsing Library,
which can be found at:

    Official repository:    http://git.spdx.org/spdx-tools-go.git
    GitHub mirror:          http://github.com/vladvelici/spdx-go
*/
package main

import (
	"github.com/spdx/tools-go/rdf"
	"github.com/spdx/tools-go/spdx"
	"github.com/spdx/tools-go/tag"
)

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const version = "pre0.0"
const execName = "spdx-go"

const helpMessage = `Usage: spdx-go <flags> <input-file>

If no input file is specified, the input is read from stdin.

The formats supported by this tool are: %s.

The format "rdf" is special, when using as input, it means rdfxml_abbrev.
When used as output, it means try to autodetect the rdf format in the input
file (using raptor "guess" parser).

One (and only one) action flag must be specified. Those are:

    -c <format> for convert
    -v for validate
    -p for pretty-print
    -help
	-version

A list of the flags supported by this tool:
`

const (
	formatRdf  = "rdf"
	formatTag  = "tag"
	formatAuto = "auto"
)

// A list of all formats supported by the tool.
var formatList = []string{
	formatRdf,
	formatTag,
	rdf.Fmt_ntriples,
	rdf.Fmt_turtle,
	rdf.Fmt_rdfxmlXmp,
	rdf.Fmt_rdfxmlAbbrev,
	rdf.Fmt_rdfxml,
	rdf.Fmt_rss,
	rdf.Fmt_atom,
	rdf.Fmt_dot,
	rdf.Fmt_jsonTriples,
	rdf.Fmt_json,
	rdf.Fmt_html,
	rdf.Fmt_nquads,
}

// Flags supported by this tool.
var (
	flagConvert       = flag.String("c", "-", "Set action to convert. Convert input file to the specified format.")
	flagValidate      = flag.Bool("v", false, "Set action to validate.")
	flagFmt           = flag.Bool("p", false, "Set action to format (pretty print). This will not necessarily ")
	flagOutput        = flag.String("o", "-", "Sets the output file. If not set, output is written to stdout.")
	flagInPlace       = flag.Bool("w", false, "If defined, it overwrites the input file.")
	flagCaseSensitive = flag.Bool("cs", false, "Case-Sensitivity of properties. Only in tag format. (if false, it treats \"packagename\" same as \"PackageName\")")
	flagInputFormat   = flag.String("f", "auto", "Defines the format of the input. Valid values: any <format> or auto. Default is auto.")
	flagHelp          = flag.Bool("help", false, "Show help message.")
	flagVersion       = flag.Bool("version", false, "Show tool version and supported SPDX spec versions.")
	flagHTML          = flag.Bool("html", false, "In validation, open a browser with visual validation results. If -o is specified, write HTML to file instead.")
)

var (
	input  = os.Stdin  // input *os.File
	output = os.Stdout // output *os.File
)

// Simple xor function.
func xor(a, b bool) bool { return a != b }

// Exits the program and prints err in an appropriate format.
func exitErr(err error) {
	switch e := err.(type) {
	default:
		log.Fatal(err)
	case *spdx.ParseError:
		if e.LineStart != e.LineEnd {
			log.Fatalf("%s:(%d to %d) %s", input.Name(), e.LineStart, e.LineEnd, e.Error())
		}
		log.Fatalf("%s:%d %s", input.Name(), e.LineStart, e.Error())
	}
}

// Program entry point. Flag processing, validation and delegation to one of the actions.
func main() {
	flag.Parse()
	log.SetFlags(0)

	if *flagHelp {
		help()
		return
	}

	if *flagVersion {
		printVersion()
		return
	}

	if !xor(*flagConvert != "-", xor(*flagValidate, *flagFmt)) {
		log.Fatal("No or invalid action flag specified. See -help for usage.")
	}

	*flagConvert = strings.ToLower(*flagConvert)
	if *flagConvert != "-" && !validFormat(*flagConvert, false) {
		log.Fatalf("No or invalid output format (-f) specified (%s). Valid values are '%s' and '%s'.", *flagConvert, formatRdf, formatTag)
	}

	if !validFormat(*flagInputFormat, true) {
		log.Fatalf("Invalid input format (-f). Valid values are '%s', '%s' and '%s'.", formatRdf, formatTag, formatAuto)
	}

	if *flagInPlace && *flagOutput != "-" {
		log.Fatal("Cannot have both -w and -o set. See -help for usage.")
	}

	if flag.NArg() >= 1 {
		var err error
		input, err = os.Open(flag.Arg(0))
		defer input.Close()
		if err != nil {
			exitErr(err)
		}
	}

	if *flagOutput != "-" {
		var err error
		output, err = os.Create(*flagOutput)
		defer output.Close()
		if err != nil {
			exitErr(err)
		}
	} else if *flagInPlace {
		if input == os.Stdin {
			log.Fatal("Cannot use -w flag when input is stdin. Please specify an input file. See -help for usage.")
		}
		var err error
		output, err = ioutil.TempFile("", "spdx-go_")
		if err != nil {
			log.Fatal(err.Error())
		}

		defer func() {
			output.Close()
			input.Close()
			CopyFile(output.Name(), input.Name())
			if err := os.Remove(output.Name()); err != nil {
				exitErr(err)
			}
		}()
	}

	// auto-detect format
	if *flagInputFormat == formatAuto {
		format := detectFormat()
		flagInputFormat = &format
	}

	if *flagConvert != "-" {
		convert()
	} else if *flagValidate {
		validate()
	} else if *flagFmt {
		format()
	}
}

// Checks whether `val` is a valid input format.
// If `allowAuto` is set to `false`, the format `"auto" (constant `formatAuto`)
// will be considered invalid.
func validFormat(val string, allowAuto bool) bool {
	if val == formatRdf || val == formatTag || (val == formatAuto && allowAuto) {
		return true
	}
	return rdf.FormatOk(val)
}

// Tries to guess the format of the input file. Does not work on stdin.
// Current method:
// 1. If input file extension is .tag or .rdf, the format is Tag or RDF, respectively.
// 2. If the file starts with <?xml, <rdf, <!-- or @import, the format is RDF, otherwise Tag
func detectFormat() string {
	if input == os.Stdin {
		log.Fatal("Cannot auto-detect format from stdin.")
	}

	if dot := strings.LastIndex(input.Name(), "."); dot+1 < len(input.Name()) {
		// check extension (if .tag or .rdf)
		format := strings.ToLower(input.Name()[dot+1:])
		if validFormat(format, false) {
			return format
		}
	}

	// Needs improvement but not a priority.
	// Only detects XML RDF or files starting with @import (turtle format) as RDF
	defer func() {
		input.Close()
		var err error
		input, err = os.Open(input.Name())
		if err != nil {
			exitErr(err)
		}
	}()
	scanner := bufio.NewScanner(input)
	scanner.Split(bufio.ScanWords)
	if scanner.Scan() {
		word := strings.ToLower(scanner.Text())
		if strings.HasPrefix(word, "<?xml") || strings.HasPrefix(word, "<rdf") || strings.HasPrefix(word, "<!--") || strings.HasPrefix(word, "@import") {
			return formatRdf
		}
		return formatTag
	}

	log.Fatal("Cannot auto-detect input format from file extension. Please use -f flag.")
	return ""
}

// Convert between SPDX formats action.
func convert() {
	var doc *spdx.Document
	var err error

	if *flagInputFormat == formatTag {
		tag.CaseSensitive(*flagCaseSensitive)
		doc, err = tag.Build(input)
	} else {
		doc, err = rdf.Parse(input, *flagInputFormat)
	}

	if err != nil {
		exitErr(err)
	}

	if *flagConvert == formatTag {
		err = tag.Write(output, doc)
	} else {
		err = rdf.WriteFormat(output, doc, *flagConvert)
	}

	if err != nil {
		exitErr(err)
	}
}

// Validate action, text outout.
func validate() {
	var doc *spdx.Document
	var err error

	if *flagInputFormat == formatTag {
		tag.CaseSensitive(*flagCaseSensitive)
		doc, err = tag.Build(input)
	} else {
		doc, err = rdf.Parse(input, *flagInputFormat)
	}

	if err != nil {
		exitErr(err)
	}

	validator := spdx.NewValidator()
	validator.Document(doc)

	if *flagHTML {
		validateHtml(doc, validator)
		return
	}

	if validator.Ok() {
		io.WriteString(output, "Document is valid.\n")
		os.Exit(0)
	}

	errs := validator.Errors()
	warnings, errors := 0, 0
	for _, e := range errs {
		if e.Type == spdx.ValidError {
			errors++
		} else {
			warnings++
		}
	}
	io.WriteString(output, fmt.Sprintf("Document is invalid. %d errors and %d warnings.\n", errors, warnings))
	for _, e := range errs {
		var meta string
		if e.Meta != nil {
			if e.Meta.LineStart == e.Meta.LineEnd {
				meta = fmt.Sprintf(":%d ", e.Meta.LineStart)
			} else {
				meta = fmt.Sprintf(":%d to %d ", e.Meta.LineStart, e.Meta.LineEnd)
			}
		} else {
			meta = " "
		}
		io.WriteString(output, input.Name()+meta+e.Error()+"\n")
	}

}

// Represents a line in the input file, used for rendering the
// HTML validation template.
type line struct {
	Number    int
	Classname string
	Content   string
	Errors    []*spdx.ValidationError
}

// The main struct used while rendering the HTML
// validation template. Keeps all the information
// shown on the page.
type summary struct {
	Lines        []*line
	FileName     string
	NoOfErrors   int
	NoOfWarnings int
	OtherErrors  []*spdx.ValidationError
}

// Validate and output HTML.
func validateHtml(doc *spdx.Document, validator *spdx.Validator) {
	sum := new(summary)
	sum.FileName = input.Name()
	sum.OtherErrors = make([]*spdx.ValidationError, 0)

	errmap := make(map[int][]*spdx.ValidationError)
	for _, e := range validator.Errors() {
		if e.Meta != nil {
			if _, ok := errmap[e.Meta.LineStart]; ok {
				errmap[e.Meta.LineStart] = append(errmap[e.Meta.LineStart], e)
			} else {
				errmap[e.Meta.LineStart] = []*spdx.ValidationError{e}
			}
		} else {
			sum.OtherErrors = append(sum.OtherErrors, e)
		}
		if e.Type == spdx.ValidWarning {
			sum.NoOfWarnings++
		} else {
			sum.NoOfErrors++
		}
	}

	// read input again
	input.Close()
	input, err := os.Open(sum.FileName)
	defer input.Close()
	if err != nil {
		exitErr(err)
	}

	sum.Lines = make([]*line, 0, 10)

	scanner := bufio.NewScanner(input)
	i := 1
	for scanner.Scan() {
		clsn := "valid"
		if len(errmap[i]) > 0 {
			clsn = "invalid"
		}
		sum.Lines = append(sum.Lines, &line{
			Number:    i,
			Content:   scanner.Text(),
			Errors:    errmap[i],
			Classname: clsn,
		})
		i++
	}
	if scanner.Err() != nil {
		exitErr(scanner.Err())
	}

	if *flagOutput == "-" {
		dir, err := ioutil.TempDir("", "spdx-go")
		if err != nil {
			exitErr(err)
		}
		output, err = os.Create(filepath.Join(dir, "spdx-go-valid.html"))
		if err != nil {
			exitErr(err)
		}
	}

	tmpl, err := template.ParseFiles("validation.html")
	if err != nil {
		exitErr(err)
	}

	err = tmpl.Execute(output, sum)
	if err != nil {
		exitErr(err)
	}

	if *flagOutput == "-" {
		if !startBrowser(output.Name()) {
			exitErr(errors.New("Couldn't open browser. Generated file: " + output.Name()))
		}
	}
}

// Credits: https://code.google.com/p/go/source/browse/cmd/cover/html.go?repo=tools
func startBrowser(url string) bool {
	// try to start the browser
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"open"}
	case "windows":
		args = []string{"cmd", "/c", "start"}
	default:
		args = []string{"xdg-open"}
	}
	cmd := exec.Command(args[0], append(args[1:], url)...)
	return cmd.Start() == nil
}

// Format action.
func format() {
	if *flagInputFormat == formatTag {
		f := tag.NewFormatter(output)
		lex := tag.NewLexer(input)
		err := f.Lexer(lex)
		if err != nil {
			exitErr(err)
		}
		return
	}

	err := rdf.WriteRdf(input, output, *flagInputFormat, *flagInputFormat)
	if err != nil {
		exitErr(err)
	}
}

// Print help message.
func help() {
	printVersion()

	fmt.Println()
	fmt.Printf(helpMessage, strings.Join(formatList, ", "))

	flag.PrintDefaults()
}

// Print tool version and supported SPDX versions.
func printVersion() {
	versions := make([]string, len(spdx.SpecVersions))
	for i, ver := range spdx.SpecVersions {
		versions[i] = fmt.Sprintf("SPDX-%d.%d", ver[0], ver[1])
	}
	fmt.Printf("spdx-go version %s.\nSupporting SPDX specifications %s.\n", version, strings.Join(versions, ", "))
}
