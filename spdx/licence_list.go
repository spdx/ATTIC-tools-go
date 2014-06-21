package spdx

import (
	"bufio"
	"os"
	"strings"
)

// Should be configured accordingly.
var LicenceListFile = "licence-list.txt"

var licenceList map[string]interface{}

func InitLicenceList() error {
	reader, err := os.Open(LicenceListFile)

	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(reader)

	licenceList = make(map[string]interface{})

	for scanner.Scan() {
		txt := strings.TrimSpace(scanner.Text())
		if txt != "" {
			licenceList[txt] = nil
		}
	}

	return scanner.Err()
}

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
