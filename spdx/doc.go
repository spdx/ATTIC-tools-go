/*
The package SPDX offers functionality to represent and manipulate SPDX
documents.

Packages spdx/tag and spdx/rdf provide functionality to parse and write SPDX
documents in and from Tag and RDF formats respectively.

The current version of the SPDX specification implemented is SPDX-1.2

For parsing documentation please refer to `spdx/tag` and `spdx/rdf` packages.

Data model
==========

Values and the medatada
-----------------------

Each SPDX element can hold the associated metadata with it. This is very useful
at validation, to provide line numbers of where the errors came from.

The interface `Value` has two methods:

    V() string // returns the actual value of the element
    M() *Meta  // returns the metadata associated with the element

All the SPDX Elements implement the `Value` interface.

Helper structs are made to store all the basic values:

    ValueStr        simple string value
    ValueBool       simple bool value
    ValueCreator    attempts to parse a "Who: What (email)" string into
                    three fields. The original value is also kept.
    ValueDate       attempts to parse a date value

SPDX Elements
-------------

The SPDX Elements are:

    Document
    CreationInfo
    Review
    Package
    File
    Licence
    ExtractedLicence

A SPDX document is stored in the `Document` struct. The `Document` is the root
of a SPDX Document and has everything: spdx version, creation information,
files, packages and extracted licences.

All SPDX Elements have a `Meta` field which stores metadata about the document.

Licences
--------

All licence structs implement the `AnyLicence` interface.

    Licence                 a licence in the SPDX licence list
    ExtractedLicence        a licence that is defined in the document, with an
                            ID that starts with "LicenseRef".
    ConjunctiveLicenceSet   a list of `AnyLicence`
    DisjunctiveLicenceSet   a list of `AnyLicence`

Validation
==========

The Validator struct is used to validate SPDX documents. It can be used to
either validate a Document, Package, CreationInfo, Review, Licence, etc.

Validating an elements fires up validation for all the nested and referenced
elements. Validating a Document also checks that all the licence and file
references are in place (everything that is used is also defined).

Licence List licences
=====================

This package uses a file (variable `LicenceListFile`) that has one licence ID
per line to check whether a licence is in the SPDX Licence List or not.

The `update-list.sh` script is provided to generate an up-to-date such list from
the official SPDX Licence List repository, which is set up as a git submodule
in this repository.

Default settings assume the licence list to be a file named `licence-list.txt`
in the root of this repository. This file is in .gitignore.

Using the script
----------------

    # in the root of this repository
    ./update-list.sh

It will initiate the submodule, pull the latest version of it and create a file
named `licence-list.txt`.
*/
package spdx
