package main

import (
	"github.com/vladvelici/spdx-go/rdf"
	"github.com/vladvelici/spdx-go/spdx"
	"github.com/vladvelici/spdx-go/tag"
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

const (
	formatRdf  = "rdf"
	formatTag  = "tag"
	formatAuto = "auto"
)

var (
	flagConvert       = flag.String("c", "-", "Convert input file to the specified format.")
	flagValidate      = flag.Bool("v", false, "Set action to validate.")
	flagFmt           = flag.Bool("p", false, "Set action to format (pretty print).")
	flagOutput        = flag.String("o", "-", "Sets the output file. If not set, output is written to stdout.")
	flagInPlace       = flag.Bool("w", false, "If defined, it overwrites the input file.")
	flagCaseSensitive = flag.Bool("cs", false, "Case-Sensitivity of properties. (if false, it treats \"packagename\" same as \"PackageName\")")
	flagInputFormat   = flag.String("f", "auto", "Defines the format of the input. Valid values: rdf, tag or auto. Default is auto.")
	flagHelp          = flag.Bool("help", false, "Show help message.")
	flagVersion       = flag.Bool("version", false, "Show tool version and supported SPDX spec versions.")
	flagHTML          = flag.Bool("html", false, "In validation, open a browser with visual validation results. If -o is specified, write HTML to file instead.")
)

var (
	input  = os.Stdin
	output = os.Stdout
)

func xor(a, b bool) bool { return a != b }

func exitErr(err error) {
	switch e := err.(type) {
	default:
		log.Fatal(err)
	case *tag.ParseError:
		if e.LineStart != e.LineEnd {
			log.Fatalf("%s:(%d to %d) %s", input.Name(), e.LineStart, e.LineEnd, e.Error())
		}
		log.Fatalf("%s:%d %s", input.Name(), e.LineStart, e.Error())
	}
}

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
		err = rdf.WriteFromat(output, doc, *flagConvert)
	}

	if err != nil {
		exitErr(err)
	}
}

func validate() {
	var doc *spdx.Document
	var err error

	if *flagInputFormat == formatTag {
		tag.CaseSensitive(*flagCaseSensitive)
		doc, err = tag.Build(input)
	} else {
		// todo: rdf
		err = errors.New("Not implemented. :(")
	}

	if err != nil {
		exitErr(err)
	}

	validator := new(spdx.Validator)
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

type line struct {
	Number    int
	Classname string
	Content   string
	Errors    []*spdx.ValidationError
}

type summary struct {
	Lines        []*line
	FileName     string
	NoOfErrors   int
	NoOfWarnings int
	OtherErrors  []*spdx.ValidationError
}

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

func format() {
	if *flagInputFormat == formatTag {
		f := tag.NewFormatter(output)
		lex := tag.NewLexer(input)
		err := f.Lexer(lex)
		if err != nil {
			exitErr(err)
		}
	}
}

func help() {
	printVersion()

	fmt.Printf("\nUsage: spdx-go [<flags>] [<input file>]\n")
	fmt.Println("Stdin is used as input if <input-file> is not specified.")

	fmt.Println("Exactly ONE of the -conv, -fmt or -valid flags MUST be specified.\n")

	flag.PrintDefaults()
}

func printVersion() {
	versions := make([]string, len(spdx.SpecVersions))
	for i, ver := range spdx.SpecVersions {
		versions[i] = fmt.Sprintf("SPDX-%d.%d", ver[0], ver[1])
	}
	fmt.Printf("spdx-go version %s.\nSupporting SPDX specifications %s.\n", version, strings.Join(versions, ", "))
}
