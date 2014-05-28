package spdx

type Review struct {
	Reviewer string // mandatory
	Date     string // mandatory
	Comment  string // optional
}
