package spdx

// File Types
const (
	FT_BINARY  = iota
	FT_SOURCE  = iota
	FT_ARCHIVE = iota
	FT_OTHER   = iota
)

// Statuses (special values such as NOASSERTION or NONE)
const (
	STATUS_NOASSERT = iota
	STATUS_NONE     = iota
	STATUS_OK       = iota
	STATUS_NOVAL    = iota
)
