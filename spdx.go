package main

import (
	"flag"
	"fmt"
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

	if flag.NArg() < 1 {
		log.Fatalf("Action not specified. For help, see '%s help' or '%s -help'", execName)
	}

	action := flag.Arg(0)

	*flagOutputFormat = strings.ToLower(*flagOutputFormat)
	if action == "convert" && !validFormat(*flagOutputFormat, false) {
		log.Fatalf("No or invalid output format (-f) specified (%s). Valid values are '%s' and '%s'.", *flagOutputFormat, formatRdf, formatTag)
	}

	if !validFormat(*flagInputFormat, true) {
		log.Fatalf("Invalid input format (-input). Valid values are '%s', '%s' and '%s'.", formatRdf, formatTag, formatAuto)
	}

	if flag.NArg() >= 2 {
		input, err := os.Open(flag.Arg(1))
		defer input.Close()
		if err != nil {
			log.Fatalf("Couldn't open input file: %s", err.Error())
		}
	}

	if *flagOutput != "-" {
		output, err := os.Create(*flagOutput)
		defer output.Close()
		if err != nil {
			log.Fatalf("Couldn't open output file: %s", err.Error())
		}
	}
}

func validFormat(val string, allowAuto bool) bool {
	return val == formatRdf || val == formatTag || (val == formatAuto && allowAuto)
}

func help() {
	printVersion()

	fmt.Printf("\nUsage: spdx-go [convert | validate | format] [<flags>] [<input file>]\n")
	fmt.Println("Stdin is used as input if <input-file> is not specified.")

	fmt.Println("\nThe following flags are available:")
	flag.PrintDefaults()

	fmt.Println("\nThe flags -help and -version do not need an action to be specified.")
}

func printVersion() {
	fmt.Printf("spdx-go version %s.\nSupporting SPDX specifications %s.\n", version, strings.Join(specVersions, ", "))
}
