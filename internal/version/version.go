// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

// Package version provides a single, shared semantic-version comparison used by
// the `update` command and the notice system, so the two never drift apart.
//
// It understands an optional leading "v"/"V", dotted numeric segments of
// arbitrary length, and a trailing pre-release suffix (e.g. "1.2.0-beta.1").
// A version carrying a pre-release suffix sorts *before* the same version
// without one ("1.2.0-rc" < "1.2.0"), matching SemVer ordering. Missing
// numeric segments are treated as zero ("1.2" == "1.2.0").
package version

import (
	"strconv"
	"strings"
)

// Compare returns -1 if a < b, 0 if a == b, and 1 if a > b.
func Compare(a, b string) int {
	aCore, aPre := split(a)
	bCore, bPre := split(b)

	if c := compareCore(aCore, bCore); c != 0 {
		return c
	}
	return comparePreRelease(aPre, bPre)
}

// split normalizes a version string and separates the numeric core from an
// optional pre-release suffix introduced by '-'. Build metadata ('+') is
// ignored for ordering purposes, per SemVer.
func split(v string) (core string, pre string) {
	v = strings.TrimSpace(v)
	if len(v) > 0 && (v[0] == 'v' || v[0] == 'V') {
		v = v[1:]
	}
	if i := strings.IndexByte(v, '+'); i >= 0 {
		v = v[:i] // drop build metadata
	}
	if i := strings.IndexByte(v, '-'); i >= 0 {
		return v[:i], v[i+1:]
	}
	return v, ""
}

// compareCore compares two dotted numeric version cores segment by segment,
// treating absent segments as zero.
func compareCore(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	n := len(aParts)
	if len(bParts) > n {
		n = len(bParts)
	}

	for i := 0; i < n; i++ {
		an := segment(aParts, i)
		bn := segment(bParts, i)
		if an < bn {
			return -1
		}
		if an > bn {
			return 1
		}
	}
	return 0
}

// segment returns the integer value of the i-th part, or 0 when out of range or
// non-numeric.
func segment(parts []string, i int) int {
	if i >= len(parts) {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(parts[i]))
	if err != nil {
		return 0
	}
	return n
}

// comparePreRelease applies SemVer pre-release ordering: a version without a
// pre-release outranks one that has it; otherwise suffixes are compared
// segment-wise (numeric segments numerically, others lexically).
func comparePreRelease(a, b string) int {
	if a == b {
		return 0
	}
	if a == "" {
		return 1 // release > pre-release
	}
	if b == "" {
		return -1
	}

	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")
	n := len(aParts)
	if len(bParts) < n {
		n = len(bParts)
	}

	for i := 0; i < n; i++ {
		if c := comparePreSegment(aParts[i], bParts[i]); c != 0 {
			return c
		}
	}
	// All shared segments equal: the longer pre-release is greater.
	switch {
	case len(aParts) < len(bParts):
		return -1
	case len(aParts) > len(bParts):
		return 1
	default:
		return 0
	}
}

// comparePreSegment compares a single pre-release identifier. Two numeric
// identifiers compare numerically; a numeric identifier is always lower than a
// non-numeric one; otherwise comparison is lexical.
func comparePreSegment(a, b string) int {
	an, aErr := strconv.Atoi(a)
	bn, bErr := strconv.Atoi(b)

	switch {
	case aErr == nil && bErr == nil:
		switch {
		case an < bn:
			return -1
		case an > bn:
			return 1
		default:
			return 0
		}
	case aErr == nil: // numeric < alphanumeric
		return -1
	case bErr == nil:
		return 1
	default:
		return strings.Compare(a, b)
	}
}
