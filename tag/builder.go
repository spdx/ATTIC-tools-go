package tag

import "github.com/vladvelici/spdx-go/spdx"

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrNoClosedParen             = errors.New("No closed parentheses at the end.")
	ErrInvalidVerifCodeExcludes  = errors.New("VerificationCode: Invalid Excludes format")
	ErrInvalidChecksum           = errors.New("Invalid Package Checksum format.")
	ErrConjunctionAndDisjunction = errors.New("Licence sets can only have either disjunction or conjunction, not both. (AND or OR, not both)")
)

var (
	orSeparator = regexp.MustCompile("(?i)\\s+or\\s+")
	andSeprator = regexp.MustCompile("(?i)\\s+and\\s+")
)

// Updater
type updater (func(string) error)
type updaterMapping map[string]updater

// Update a string pointer
func upd(ptr *string) updater {
	return func(val string) error {
		*ptr = val
		return nil
	}
}

// Update a string slice pointer
func updList(arr *[]string) updater {
	return func(val string) error {
		*arr = append(*arr, val)
		return nil
	}
}

// Update a VerificationCode pointer
func verifCode(vc *spdx.VerificationCode) updater {
	return func(val string) error {
		vc.Value = val
		open := strings.Index(val, "(")
		if open > 0 {
			vc.Value = strings.TrimSpace(val[:open])

			val = val[open+1:]

			// close parentheses
			cls := strings.LastIndex(val, ")")
			if cls < 0 {
				return ErrNoClosedParen
			}

			val = val[:cls]

			// remove "excludes:" if exists
			excludeRegexp := regexp.MustCompile("(?i)excludes:\\s*")
			val = excludeRegexp.ReplaceAllLiteralString(val, "")

			vc.ExcludedFiles = strings.Split(val, ",")
			for i, _ := range vc.ExcludedFiles {
				vc.ExcludedFiles[i] = strings.TrimSpace(vc.ExcludedFiles[i])
			}
		}
		return nil
	}
}

// Update a Checksum pointer
func checksum(cksum *spdx.Checksum) updater {
	return func(val string) error {
		split := strings.Split(val, ":")
		if len(split) != 2 {
			return ErrInvalidChecksum
		}
		cksum.Algo, cksum.Value = strings.TrimSpace(split[0]), strings.TrimSpace(split[1])
		return nil
	}
}

// Finds the bigger set of open-close parantheses.
// If there is no open parentheses it returns -1 and -2.
// If there is no closing parantheses for the first open parantheses found,
// it returns open to be the index where the first open parantheses is found
// and close to be -2.
//
// To check whether a good match has been found, check whether open < close rather than open >= 0
//
// Example:
//  "a and (b or (c and d))"
//  returns: open=7, cls=21 (input[open:cls+1]=="(b or (c and d))")
func findMatchingParenSet(str string) (open, cls int) {
	open, cls = -1, -2
	open = strings.Index(str, "(")
	if open < 0 {
		return
	}
	count := 0
	for i := open; i < len(str); i++ {
		if str[i] == '(' {
			count++
		} else if str[i] == ')' {
			count--
		}
		if count == 0 {
			return open, i
		}
	}
	return
}

// Determines if a licence set string is disjunctive, conjunctive, both or none.
// Assumes balanced parantheses in the input.
func conjOrDisjSet(str string) (conj, disj bool) {
	str = strings.TrimSpace(str)
	if str == "" {
		return
	}

	// clear parentheses
	for open, cls := findMatchingParenSet(str); cls > open; open, cls = findMatchingParenSet(str) {
		if cls == len(str)-1 {
			str = str[:open]
		} else {
			str = str[:open] + str[cls+1:]
		}
	}

	// test both and and or separators
	conj = andSeprator.FindStringIndex(str) != nil
	disj = orSeparator.FindStringIndex(str) != nil

	return
}

// Splits a licence set string into tokens separated by the given separator.
// Ignores the separator if it is contained in parentheses.
func licenceSetSplit(sep *regexp.Regexp, str string) []string {
	separators := sep.FindAllStringIndex(str, -1)
	if separators == nil {
		return []string{str}
	}

	used := 0
	result := make([]string, 0)
	for i := 0; i < len(separators); i++ {
		nextOpen, nextClose := findMatchingParenSet(str[used:])
		if nextOpen >= 0 && nextClose >= 0 {
			nextOpen += used
			nextClose += used
		}

		nextSep := separators[i]
		if nextOpen < nextSep[0] && nextSep[1] < nextClose {
			// find a new token that's after nextClose
			continue
		}

		result = append(result, strings.TrimSpace(str[used:nextSep[0]]))
		used = nextSep[1]
	}

	lastToken := strings.TrimSpace(str[used:])
	if len(lastToken) > 0 {
		result = append(result, lastToken)
	}

	return result
}

// Parses sets of licences
func parseLicenceSet(val string) (spdx.AnyLicenceInfo, error) {
	val = strings.TrimSpace(val)
	if len(val) == 0 {
		return nil, nil
	}

	// if everything is in parentheses, remove the big parentheses
	o, c := findMatchingParenSet(val)
	if o == 0 && c == len(val)-1 {
		if len(val) <= 2 {
			return nil, errors.New("Empty licence identifier")
		}
		val = val[1 : len(val)-1]
	}

	conj, disj := conjOrDisjSet(val)

	if disj && conj {
		return nil, ErrConjunctionAndDisjunction
	}

	if conj {
		tokens := licenceSetSplit(andSeprator, val)
		res := make(spdx.ConjunctiveLicenceList, len(tokens))
		for _, t := range tokens {
			lic, err := parseLicenceSet(t)
			if err != nil {
				return nil, err
			}
			res = append(res, lic)
		}
		return res, nil
	}

	if disj {
		tokens := licenceSetSplit(orSeparator, val)
		res := make(spdx.DisjunctiveLicenceList, len(tokens))
		for _, t := range tokens {
			lic, err := parseLicenceSet(t)
			if err != nil {
				return nil, err
			}
			res = append(res, lic)
		}
		return res, nil
	}

	return spdx.NewLicenceReference(strings.TrimSpace(val)), nil

}

// Given a value from the pair, returns the appropriate spdx.AnyLicenceInfo
func parseLicenceString(val string) (spdx.AnyLicenceInfo, error) {
	if len(val) == 0 {
		return nil, nil
	}
	openParen := strings.Count(val, "(")
	closeParen := strings.Count(val, ")")

	if openParen != closeParen {
		return nil, ErrNoClosedParen
	}

	if openParen > 0 {
		return parseLicenceSet(val)
	}

	return spdx.NewLicenceReference(strings.TrimSpace(val)), nil
}

// Update a AnyLicenceInfo pointer
func anyLicence(lic *spdx.AnyLicenceInfo) updater {
	return func(val string) error {
		return nil
	}
}

// Update a []anyLicenceInfo pointer
func anyLicenceList(licList *[]spdx.AnyLicenceInfo) updater {
	return func(val string) error {
		return nil
	}
}

// Creates a file that only has the FileName and appends it to the initially given pointer
func updFileNameList(fl *[]*spdx.File) updater {
	return func(val string) error {
		file := &spdx.File{Name: val}
		*fl = append(*fl, file)
		return nil
	}
}

// Gets all the key/value combinations in src and puts them in dest (overwrites if values already exist)
func mapMerge(dest *updaterMapping, src updaterMapping) {
	mp := *dest
	for k, v := range src {
		mp[k] = v
	}
}

// Document mapping.
func documentMap(doc *spdx.Document) updaterMapping {
	doc.CreationInfo = new(spdx.CreationInfo)
	doc.CreationInfo.Creator = make([]string, 0)
	doc.Packages = make([]*spdx.Package, 0)
	doc.Files = make([]*spdx.File, 0)
	doc.ExtractedLicenceInfo = make([]*spdx.ExtractedLicensingInfo, 0)

	var mapping updaterMapping

	mapping = map[string]updater{
		// SpdxDocument
		"SpecVersion":        upd(&doc.SpecVersion),
		"DataLicense":        upd(&doc.DataLicence),
		"DocumentComment":    upd(&doc.Comment),
		"Creator":            updList(&doc.CreationInfo.Creator),
		"Created":            upd(&doc.CreationInfo.Created),
		"CreatorComment":     upd(&doc.CreationInfo.Comment),
		"LicenseListVersion": upd(&doc.CreationInfo.LicenceListVersion),

		// Package
		"PackageName": func(val string) error {
			pkg := &spdx.Package{
				Name:             val,
				Checksum:         new(spdx.Checksum),
				VerificationCode: new(spdx.VerificationCode),
			}
			doc.Packages = append(doc.Packages, pkg)

			// Add package values that are now available
			mapMerge(&mapping, updaterMapping{
				"PackageVersion":              upd(&pkg.Version),
				"PackageFileName":             upd(&pkg.FileName),
				"PackageSupplier":             upd(&pkg.Supplier),
				"PackageOriginator":           upd(&pkg.Originator),
				"PackageDownloadLocation":     upd(&pkg.DownloadLocation),
				"PackageVerificationCode":     verifCode(pkg.VerificationCode),
				"PackageChecksum":             checksum(pkg.Checksum),
				"PackageHomePage":             upd(&pkg.HomePage),
				"PackageSourceInfo":           upd(&pkg.SourceInfo),
				"PackageLicenseConcluded":     anyLicence(&pkg.LicenceConcluded),
				"PackageLicenseInfoFromFiles": anyLicenceList(&pkg.LicenceInfoFromFiles),
				"PackageLicenseDeclared":      anyLicence(&pkg.LicenceDeclared),
				"PackageLicenseComments":      upd(&pkg.LicenceComments),
				"PackageCopyrightText":        upd(&pkg.CopyrightText),
				"PackageSummary":              upd(&pkg.Summary),
				"PackageDescription":          upd(&pkg.Description),
			})

			return nil
		},
		// File
		"FileName": func(val string) error {
			file := &spdx.File{
				Checksum:   new(spdx.Checksum),
				Dependency: make([]*spdx.File, 0),
				ArtifactOf: make([]*spdx.ArtifactOf, 0),
			}

			mapMerge(&mapping, updaterMapping{
				"FileType":          upd(&file.Type),
				"FileChecksum":      checksum(file.Checksum),
				"LicenseConcluded":  anyLicence(&file.LicenceConcluded),
				"LicenseInfoInFile": anyLicenceList(&file.LicenceInfoInFile),
				"LicenseComments":   upd(&file.LicenceComments),
				"FileCopyrightText": upd(&file.CopyrightText),
				"FileComment":       upd(&file.Comment),
				"FileNotice":        upd(&file.Notice),
				"FileContributor":   updList(&file.Contributor),
				"FileDependency":    updFileNameList(&file.Dependency),
				"ArtifactOfProjectName": func(val string) error {
					artif := new(spdx.ArtifactOf)
					mapMerge(&mapping, updaterMapping{
						"ArtifactOfProjectHomepage": upd(&artif.HomePage),
						"ArtifactOfProjectUri":      upd(&artif.ProjectUri),
					})
					file.ArtifactOf = append(file.ArtifactOf, artif)
					return nil
				},
			})

			return nil
		},

		// ExtractedLicensingInfo
		"LicenseID": func(val string) error {
			lic := &spdx.ExtractedLicensingInfo{
				Name:           make([]string, 0),
				CrossReference: make([]string, 0),
			}
			mapMerge(&mapping, updaterMapping{
				"ExtractedText":         upd(&lic.Text),
				"LicenseName":           updList(&lic.Name),
				"LicenseCrossReference": updList(&lic.CrossReference),
				"LicenseComment":        upd(&lic.Comment),
			})
			doc.ExtractedLicenceInfo = append(doc.ExtractedLicenceInfo, lic)
			return nil
		},
	}

	return mapping
}

// Apply the relevant updater function if the given pair matches any.
//
// ok means whether the property was in the map or not
// err is the error returned by applying the mapping function or, if ok == false, an error with the relevant "mapping not found" message
//
// It returns two arguments to allow for easily creating parsing modes such as "ignore not known mapping"
func applyMapping(input Pair, mapping updaterMapping) (ok bool, err error) {
	f, ok := mapping[input.Key]
	if !ok {
		return false, errors.New("Invalid property or property needs another property to be defined before it: " + input.Key)
	}
	return true, f(input.Value)
}

// Parse a []Pair list to a *spdx.Document
func parseDocument(input []Pair) (*spdx.Document, error) {
	doc := new(spdx.Document)
	mapping := documentMap(doc)

	for _, p := range input {
		_, err := applyMapping(p, mapping)
		if err != nil {
			return nil, err
		}
	}

	return doc, nil
}
