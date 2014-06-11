package main

import (
	"github.com/vladvelici/spdx-go/spdx"
	"github.com/vladvelici/spdx-go/tag"
)

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const version = "pre0.0"
const execName = "spdx-go"

var specVersions = []string{"SPDX-1.2"}

const (
	formatRdf  = "rdf"
	formatTag  = "tag"
	formatAuto = "auto"
)

var (
	flagConvert     = flag.String("c", "-", "Convert input file to the specified format.")
	flagValidate    = flag.Bool("v", false, "Set action to validate.")
	flagFmt         = flag.Bool("p", false, "Set action to format (pretty print).")
	flagOutput      = flag.String("o", "-", "Sets the output file. If not set, output is written to stdout.")
	flagInPlace     = flag.Bool("w", false, "If defined, it overwrites the input file.")
	flagIgnoreCase  = flag.Bool("i", false, "If defined, it ignores the case for properties. (e.g. treat \"packagename\" same as \"PackageName\")")
	flagInputFormat = flag.String("f", "auto", "Defines the format of the input. Valid values: rdf, tag or auto. Default is auto.")
	flagHelp        = flag.Bool("help", false, "Show help message.")
	flagVersion     = flag.Bool("version", false, "Show tool version and supported SPDX spec versions.")
)

var (
	input  = os.Stdin
	output = os.Stdout
)

func xor(a, b bool) bool { return a != b }

func main() {
	flag.Parse()

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
	if !validFormat(*flagConvert, false) {
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
			log.Fatal(err.Error())
		}
	}

	if *flagOutput != "-" {
		var err error
		output, err = os.Create(*flagOutput)
		defer output.Close()
		if err != nil {
			log.Fatal(err.Error())
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
				log.Fatal(err.Error())
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
	return val == formatRdf || val == formatTag || (val == formatAuto && allowAuto)
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
			log.Fatal(err.Error())
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
		doc, err = tag.Build(input)
	} else {
		// doc, err = rdf.Parse(input)
	}

	if err != nil {
		log.Fatal(err)
	}

	if *flagConvert == formatTag {
		tag.Write(output, doc)
	} else {
		// rdf write
	}
}

func validate() {
}

func format() {
}

func help() {
	printVersion()

	fmt.Printf("\nUsage: spdx-go [<flags>] [<input file>]\n")
	fmt.Println("Stdin is used as input if <input-file> is not specified.")

	fmt.Println("Exactly ONE of the -conv, -fmt or -valid flags MUST be specified.\n")

	flag.PrintDefaults()
}

func printVersion() {
	fmt.Printf("spdx-go version %s.\nSupporting SPDX specifications %s.\n", version, strings.Join(specVersions, ", "))
}
