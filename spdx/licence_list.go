package spdx

import (
	"bufio"
	"os"
	"strings"
)

// Should be configured by clients so that it reflects the location of a SPDX
// licence list. The file must have a licence ID per line. Empty lines and
// spaces are ignored.
//
// The script ../update-list.sh can be used to generate the list.
var LicenceListFile = "licence-list.txt"

// Set for looking up licence IDs. Do not use directly, use CheckLicence()
// instead.
var licenceList map[string]interface{}

// Initialises the licenceList set. Call before using CheckLicence as it returns
// the IO error for reading the LicenceListFile if there is any. CheckLicence()
// will initialise the licenceList set if not already initialised, but it will
// panic if IO errors occour.
func InitLicenceList() error {
	licenceList = make(map[string]interface{})

	reader, err := os.Open(LicenceListFile)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		txt := strings.TrimSpace(scanner.Text())
		if txt != "" {
			licenceList[txt] = nil
		}
	}

	return scanner.Err()
}

// CheckLicence checks whether the licence ID `lic` is in the SPDX Licence List.
// Calls InitLicenceList() if has not been called before and, if it returns an
// error, it panics with that error.
func CheckLicence(lic string) bool {
	if licenceList == nil {
		err := InitLicenceList()
		if err != nil {
			panic(err)
		}
	}
	_, ok := licenceList[lic]
	return ok
}
