package main

import (
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
	flagConvert      = flag.Bool("conv", false, "Set action to convert.")
	flagValidate     = flag.Bool("valid", false, "Set action to validate.")
	flagFmt          = flag.Bool("fmt", false, "Set action to format (pretty print).")
	flagOutput       = flag.String("o", "-", "Sets the output file. If not set, output is written to stdout.")
	flagInPlace      = flag.Bool("w", false, "If defined, it overwrites the input file.")
	flagIgnoreCase   = flag.Bool("i", false, "If defined, it ignores the case for properties. (e.g. treat \"packagename\" same as \"PackageName\")")
	flagInputFormat  = flag.String("input", "auto", "Defines the format of the input. Valid values: rdf, tag or auto. Default is auto.")
	flagOutputFormat = flag.String("f", "", "Mandatory on convert action, it specifies the output format (rdf or tag).")
	flagHelp         = flag.Bool("help", false, "Show help message.")
	flagVersion      = flag.Bool("version", false, "Show tool version and supported SPDX spec versions.")
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

	if !xor(*flagConvert, xor(*flagValidate, *flagFmt)) {
		log.Fatal("No or invalid action flag specified. See -help for usage.")
	}

	*flagOutputFormat = strings.ToLower(*flagOutputFormat)
	if *flagConvert && !validFormat(*flagOutputFormat, false) {
		log.Fatalf("No or invalid output format (-f) specified (%s). Valid values are '%s' and '%s'.", *flagOutputFormat, formatRdf, formatTag)
	}

	if !validFormat(*flagInputFormat, true) {
		log.Fatalf("Invalid input format (-input). Valid values are '%s', '%s' and '%s'.", formatRdf, formatTag, formatAuto)
	}

	if *flagInPlace && *flagOutput != "-" {
		log.Fatal("Cannot have both -w and -o set. See -help for usage.")
	}

	if flag.NArg() >= 1 {
		input, err := os.Open(flag.Arg(0))
		defer input.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	if *flagOutput != "-" {
		output, err := os.Create(*flagOutput)
		defer output.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	} else if *flagInPlace {
		if input == os.Stdin {
			log.Fatal("Cannot use -w flag when input is stdin. Please specify an input file. See -help for usage.")
		}

		output, err := ioutil.TempFile("", "spdx-go_")
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
		flagInputFormat = detectFormat()
	}

	if *flagConvert {
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

// Currently only detects .rdf and .tag extensions.
func detectFormat() *string {
	dot := strings.LastIndex(input.Name(), ".")
	if dot < 0 || dot+1 == len(input.Name()) {
		log.Fatal("Cannot auto-detect input format. Please specify format using the -input flag.")
	}

	// check extension (if .tag or .rdf)
	format := strings.ToLower(input.Name()[dot+1:])
	if validFormat(format, false) {
		return &format
	}

	// TODO: try to detect format by scanning file header

	log.Fatal("Cannot auto-detect input format from file extension. Please use -input flag.")
	return nil
}

func convert() {

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
