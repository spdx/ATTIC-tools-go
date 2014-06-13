package tag

import "github.com/vladvelici/spdx-go/spdx"

import (
	"regexp"
	"strings"
)

// Error messages used in this file
var (
	MsgNoClosedParen             = "No closed parentheses at the end."
	MsgInvalidVerifCodeExcludes  = "VerificationCode: Invalid Excludes format"
	MsgInvalidChecksum           = "Invalid Package Checksum format."
	MsgConjunctionAndDisjunction = "Licence sets can only have either disjunction or conjunction, not both. (AND or OR, not both)"
	MsgEmptyLicence              = "Empty licence"
	MsgAlreadyDefined            = "Property already defined"
)

// Regular expressions to match licence set separators
var (
	orSeparator = regexp.MustCompile("(?i)\\s+or\\s+")  // disjunctive licence set separator
	andSeprator = regexp.MustCompile("(?i)\\s+and\\s+") // conjunctive licence set separator
)

// A function that takes a *Token and updates some value in a SPDX element (document, package, ...).
type updater (func(*Token) error)

// A map of SPDX Tags (properties) and updater functions
type updaterMapping map[string]updater

// Update a ValString pointer
func upd(ptr *spdx.ValueStr) updater {
	set := false
	return func(tok *Token) error {
		if set {
			return parseError(MsgAlreadyDefined, tok.Meta)
		}
		ptr.Val = tok.Pair.Value
		ptr.Meta = tok.Meta
		set = true
		return nil
	}
}

// Update a []ValString pointer
func updList(arr *[]spdx.ValueStr) updater {
	return func(tok *Token) error {
		*arr = append(*arr, spdx.Str(tok.Pair.Value, tok.Meta))
		return nil
	}
}

// Update a VerificationCode pointer
func verifCode(vc *spdx.VerificationCode) updater {
	set := false
	return func(tok *Token) error {
		if set {
			return parseError(MsgAlreadyDefined, tok.Meta)
		}
		val := tok.Pair.Value
		open := strings.Index(val, "(")
		if open > 0 {
			vc.Value = spdx.Str(strings.TrimSpace(val[:open]), tok.Meta)

			val = val[open+1:]

			// close parentheses
			cls := strings.LastIndex(val, ")")
			if cls < 0 {
				return parseError(MsgNoClosedParen, tok.Meta)
			}

			val = val[:cls]

			// remove "excludes:" if exists
			excludeRegexp := regexp.MustCompile("(?i)excludes:\\s*")
			val = excludeRegexp.ReplaceAllLiteralString(val, "")

			exclFiles := strings.Split(val, ",")
			vc.ExcludedFiles = make([]spdx.ValueStr, len(exclFiles))
			for i, v := range exclFiles {
				vc.ExcludedFiles[i] = spdx.Str(strings.TrimSpace(v), tok.Meta)
			}
		} else {
			vc.Value = spdx.Str(strings.TrimSpace(val), tok.Meta)
		}
		set = true
		return nil
	}
}

// Update a Checksum pointer
func checksum(cksum *spdx.Checksum) updater {
	set := false
	return func(tok *Token) error {
		if set {
			return parseError(MsgAlreadyDefined, tok.Meta)
		}
		split := strings.Split(tok.Pair.Value, ":")
		if len(split) != 2 {
			return parseError(MsgInvalidChecksum, tok.Meta)
		}
		cksum.Algo, cksum.Value = spdx.Str(strings.TrimSpace(split[0]), tok.Meta), spdx.Str(strings.TrimSpace(split[1]), tok.Meta)
		set = true
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

// Parses sets of licences. Assumes the input tok.Value to have balanced parentheses.
func parseLicenceSet(tok *Token) (spdx.AnyLicenceInfo, error) {
	val := strings.TrimSpace(tok.Pair.Value)
	if len(val) == 0 {
		return nil, parseError(MsgEmptyLicence, tok.Meta)
	}

	// if everything is in parentheses, remove the big parentheses
	o, c := findMatchingParenSet(val)
	if o == 0 && c == len(val)-1 {
		if len(val) <= 2 {
			return nil, parseError(MsgEmptyLicence, tok.Meta)
		}
		val = val[1 : len(val)-1]
	}

	conj, disj := conjOrDisjSet(val)

	if disj && conj {
		return nil, parseError(MsgConjunctionAndDisjunction, tok.Meta)
	}

	if conj {
		tokens := licenceSetSplit(andSeprator, val)
		res := make(spdx.ConjunctiveLicenceList, 0, len(tokens))
		for _, t := range tokens {
			lic, err := parseLicenceSet(&Token{Type: tok.Type, Meta: tok.Meta, Pair: Pair{Value: t}})
			if err != nil {
				return nil, err
			}
			res = append(res, lic)
		}
		return res, nil
	}

	if disj {
		tokens := licenceSetSplit(orSeparator, val)
		res := make(spdx.DisjunctiveLicenceList, 0, len(tokens))
		for _, t := range tokens {
			lic, err := parseLicenceSet(&Token{Type: tok.Type, Meta: tok.Meta, Pair: Pair{Value: t}})
			if err != nil {
				return nil, err
			}
			res = append(res, lic)
		}
		return res, nil
	}

	return spdx.NewLicenceReference(strings.TrimSpace(val), tok.Meta), nil

}

// Given a Token, returns the appropriate spdx.AnyLicenceInfo. If there are parentheses, it
// checks whether they are balanced and calls parseLicenceSet()
func parseLicenceString(tok *Token) (spdx.AnyLicenceInfo, error) {
	val := strings.TrimSpace(tok.Pair.Value)
	if len(val) == 0 {
		return nil, parseError(MsgEmptyLicence, tok.Meta)
	}
	openParen := strings.Count(val, "(")
	closeParen := strings.Count(val, ")")

	if openParen != closeParen {
		return nil, parseError(MsgNoClosedParen, tok.Meta)
	}

	if openParen > 0 {
		return parseLicenceSet(tok)
	}

	return spdx.NewLicenceReference(strings.TrimSpace(val), tok.Meta), nil
}

// Update a AnyLicenceInfo pointer.
func anyLicence(lic *spdx.AnyLicenceInfo) updater {
	set := false
	return func(tok *Token) error {
		if set {
			return parseError(MsgAlreadyDefined, tok.Meta)
		}
		l, err := parseLicenceString(tok)
		if err != nil {
			return err
		}
		*lic = l
		set = true
		return nil
	}
}

// Update a []anyLicenceInfo pointer
func anyLicenceList(licList *[]spdx.AnyLicenceInfo) updater {
	return func(tok *Token) error {
		l, err := parseLicenceString(tok)
		if err != nil {
			return err
		}
		*licList = append(*licList, l)
		return nil
	}
}

// Creates a file that only has the FileName and appends it to the given []*File pointer
func updFileNameList(fl *[]*spdx.File) updater {
	return func(tok *Token) error {
		file := &spdx.File{Name: spdx.Str(tok.Value, tok.Meta)}
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

// Creates the mapping of a *spdx.Document in Tag format.
func documentMap(doc *spdx.Document) updaterMapping {
	doc.CreationInfo = new(spdx.CreationInfo)
	doc.CreationInfo.Creator = make([]spdx.ValueStr, 0)

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
		"PackageName": func(tok *Token) error {
			pkg := &spdx.Package{
				Name:             spdx.Str(tok.Value, tok.Meta),
				Checksum:         new(spdx.Checksum),
				VerificationCode: new(spdx.VerificationCode),
			}

			if doc.Packages == nil {
				doc.Packages = []*spdx.Package{pkg}
			} else {
				doc.Packages = append(doc.Packages, pkg)
			}

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
		"FileName": func(tok *Token) error {
			file := &spdx.File{
				Checksum:   new(spdx.Checksum),
				Dependency: make([]*spdx.File, 0),
				ArtifactOf: make([]*spdx.ArtifactOf, 0),
				Name:       spdx.Str(tok.Value, tok.Meta),
			}

			if doc.Files == nil {
				doc.Files = []*spdx.File{file}
			} else {
				doc.Files = append(doc.Files, file)
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
				"ArtifactOfProjectName": func(tok *Token) error {
					artif := new(spdx.ArtifactOf)
					artif.Name = spdx.Str(tok.Value, tok.Meta)
					mapMerge(&mapping, updaterMapping{
						"ArtifactOfProjectHomePage": upd(&artif.HomePage),
						"ArtifactOfProjectURI":      upd(&artif.ProjectUri),
					})
					file.ArtifactOf = append(file.ArtifactOf, artif)
					return nil
				},
			})

			return nil
		},

		// ExtractedLicensingInfo
		"LicenseID": func(tok *Token) error {
			lic := &spdx.ExtractedLicensingInfo{
				Id:             spdx.Str(tok.Value, tok.Meta),
				Name:           make([]spdx.ValueStr, 0),
				CrossReference: make([]spdx.ValueStr, 0),
			}
			mapMerge(&mapping, updaterMapping{
				"ExtractedText":         upd(&lic.Text),
				"LicenseName":           updList(&lic.Name),
				"LicenseCrossReference": updList(&lic.CrossReference),
				"LicenseComment":        upd(&lic.Comment),
			})

			if doc.ExtractedLicenceInfo == nil {
				doc.ExtractedLicenceInfo = []*spdx.ExtractedLicensingInfo{lic}
			} else {
				doc.ExtractedLicenceInfo = append(doc.ExtractedLicenceInfo, lic)
			}

			return nil
		},

		"Reviewer": func(tok *Token) error {
			rev := &spdx.Review{
				Reviewer: spdx.Str(tok.Value, tok.Meta),
			}

			if doc.Reviews == nil {
				doc.Reviews = []*spdx.Review{rev}
			} else {
				doc.Reviews = append(doc.Reviews, rev)
			}

			mapMerge(&mapping, updaterMapping{
				"ReviewDate":    upd(&rev.Date),
				"ReviewComment": upd(&rev.Comment),
			})

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
func applyMapping(tok *Token, mapping updaterMapping) (ok bool, err error) {
	f, ok := mapping[tok.Key]
	if !ok {
		return false, parseError("Invalid property or property needs another property to be defined before it: "+tok.Key, tok.Meta)
	}
	return true, f(tok)
}

// Parse Tokens given by a lexer to a *spdx.Document.
// Errors returned are either I/O errors returned by the io.Reader associated with the given lexer,
// lexing errors (still have *ParseError type) or parse errors (type *ParseError).
func Parse(lex lexer) (*spdx.Document, error) {
	doc := new(spdx.Document)
	mapping := documentMap(doc)
	for lex.Lex() {
		token := lex.Token()

		// ignore comments if they're returned by lexer
		if token.Type != TokenPair {
			continue
		}

		_, err := applyMapping(token, mapping)
		if err != nil {
			return nil, err
		}
	}

	if lex.Err() != nil {
		return nil, lex.Err()
	}

	return doc, nil
}
