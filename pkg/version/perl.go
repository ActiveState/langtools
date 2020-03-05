package version

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// Ensures that all produced sortable versions have a minimum of 3
	// segments, i.e. ["1", "2", "3"]
	minimumPerlVersionSegmentCount = 3

	// These regular expression patterns are based off the lax regular
	// expressions in version/regex.pm
	// (https://metacpan.org/source/JPEACOCK/version-0.9924/lib/version/regex.pm).
	// We use lax instead of strict to accept as many version strings as
	// reasonably possible.
	integerPart  = `[0-9]+`
	fractionPart = `\.[0-9]+`
	alphaPart    = `(_[0-9]+)?`

	// decimalPattern matches the decimal version string type as defined by
	// version.pm. A decimal version does not start with a 'v' and contains 0
	// or 1 decimal points (periods). It may have a trailing underscore
	// followed by some number of digits. decimalPattern matches the following
	// examples: '1', '1.', '.2', '1.2', '1.002003', '1.002_003'
	decimalPattern = integerPart + `\.?` +
		`|` + integerPart + fractionPart + alphaPart +
		`|` + fractionPart + alphaPart

	// dottedDecimalPattern matches the dotted-decimal version string type as
	// defined by version.pm. A dotted-decimal is a version string that either
	// starts with a 'v' or contains 2 or more decimal points (periods). It
	// may have a trailing underscore followed by some number of
	// digits. dottedDecimalPattern matches the following examples: v1, v1.,
	// v1.0, v1.2.3, .2.3, 1.2.3, v1.23_456, 1.2.3.4.5_6789
	dottedDecimalPattern = `v` + integerPart + `\.?` +
		`|v` + integerPart + `(` + fractionPart + `)+` + alphaPart +
		`|(` + integerPart + `)?` + `(` + fractionPart + `){2,}` + alphaPart
)

var (
	decimalRegex       = regexp.MustCompile(`^(` + decimalPattern + `)$`)
	dottedDecimalRegex = regexp.MustCompile(`^(` + dottedDecimalPattern + `)$`)
)

// ParsePerl parses version using the version parsing algorithm used by
// version.pm (https://metacpan.org/pod/distribution/version/lib/version.pm).
// version.pm considers there to be two perl version types: decimal (1.20) and
// dotted-decimal (v1.2.3). This function parses both types and normalizes
// them to dotted-decimal for comparison purposes.
func ParsePerl(version string) (*Version, error) {
	if decimalRegex.MatchString(version) {
		return parsePerlDecimalVersion(version)
	}

	if dottedDecimalRegex.MatchString(version) {
		return parsePerlVStringVersion(version)
	}

	return nil, fmt.Errorf("not valid perl version: %s", version)
}

func parsePerlDecimalVersion(version string) (*Version, error) {
	version = strings.ReplaceAll(version, "_", "")
	parts := strings.Split(version, ".")
	segments := make([]string, 0, minimumPerlVersionSegmentCount)
	segments = append(segments, decimalIntegerPartToSegment(parts[0]))
	if len(parts) == 2 {
		segments = append(segments, decimalFractionAndAlphaPartToSegments(parts[1])...)
	}
	segments = expandToMinimumSegmentCount(segments)

	return fromStringSlice(PerlDecimal, version, segments)
}

func decimalIntegerPartToSegment(part string) string {
	// There was no integer part in the given version string, use 0.
	if part == "" {
		return "0"
	}
	return part
}

func decimalFractionAndAlphaPartToSegments(part string) []string {
	// Pad part out to a multiple of three.
	if padN := len(part) % 3; padN != 0 {
		part += "000"[padN:]
	}

	// Split part into three-digit long segments.
	segments := make([]string, 0, 2)
	for i := 0; i < len(part); i = i + 3 {
		j := min(i+3, len(part))
		segments = append(segments, removeLeadingZeros(part[i:j]))
	}

	return segments
}

func min(lhs, rhs int) int {
	if lhs <= rhs {
		return lhs
	}
	return rhs
}

func removeLeadingZeros(s string) string {
	if s = strings.TrimLeft(s, "0"); s != "" {
		return s
	}
	return "0"
}

func parsePerlVStringVersion(version string) (*Version, error) {
	version = strings.TrimPrefix(version, "v")
	version = strings.ReplaceAll(version, "_", "")
	segments := strings.Split(version, ".")
	for i, s := range segments {
		if s == "" {
			segments[i] = "0"
		}
	}
	segments = expandToMinimumSegmentCount(segments)

	return fromStringSlice(PerlVString, version, segments)
}

func expandToMinimumSegmentCount(segments []string) []string {
	for i := len(segments); i < minimumPerlVersionSegmentCount; i++ {
		segments = append(segments, "0")
	}
	return segments
}
