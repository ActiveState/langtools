package version

// This file contains parsers for version types that are not language
// specific.

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/text/unicode/norm"
)

const (
	// Value greater than unicode's upper limit of 0x10FFFF = 1,114,111
	maxValue            = "2000000"
	delimiter           = "-"
	delimitedSubsection = delimiter + "$1" + delimiter
)

var (
	// See https://github.com/google/re2/wiki/Syntax for go regex character classes.
	// \pZ  = unicode separator character class
	// \pP  = unicode punctuation character class
	anyPunctuationOrSeparator = regexp.MustCompile(`[\p{P}\p{Z}]+`)
	wholeNumber               = regexp.MustCompile(`([0-9]+)`)
	decimalNumber             = regexp.MustCompile(`^(\d+\.\d*|\.?\d+)$`)
	notZero                   = regexp.MustCompile(`[^0]`)

	// Matches semver 2.0
	semVerRegEx = regexp.MustCompile(`^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

	genericPreReleaseIdentifiers = map[string]string{
		"alpha":   "-26",
		"beta":    "-25",
		"gamma":   "-24",
		"delta":   "-23",
		"epsilon": "-22",
		"zeta":    "-21",
		"eta":     "-20",
		"theta":   "-19",
		"iota":    "-18",
		"kappa":   "-17",
		"lambda":  "-16",
		"mu":      "-15",
		"nu":      "-14",
		"xi":      "-13",
		"omicron": "-12",
		"pi":      "-11",
		"rho":     "-10",
		"sigma":   "-9",
		"tau":     "-8",
		"upsilon": "-7",
		"phi":     "-6",
		"chi":     "-5",
		"psi":     "-4",
		"omega":   "-3",
		"pre":     "-2",
		"rc":      "-1",
	}
)

// ParseGeneric parses the version string into an array of decimal numbers
// such that two parsed version strings can be compared. This function treats
// numbers as individually comparable segments and not as decimal numbers,
// i.e. 1.2 is parsed to be compared as two numbers: 1 and 2.
func ParseGeneric(version string) (*Version, error) {
	version = normalizeUnicode(version)
	segments := parseBySeparator(
		version,
		anyPunctuationOrSeparator,
		toDecimalStringWithGenericPreReleaseIdentifierHandling,
	)

	if !containsGenericPreReleaseIdentifierValue(segments) {
		segments = append(segments, "0")
	}

	return fromStringSlice(Generic, version, segments)
}

// ParseSemVer parses the semantic version (https://semver.org/) version
// string into an array of decimal numbers such that two parsed version
// strings can be compared as required by the semantic versioning
// specification.
func ParseSemVer(version string) (*Version, error) {
	matches := semVerRegEx.FindStringSubmatch(version)
	if len(matches) == 0 {
		return nil, fmt.Errorf("Version does not match semver regex: %s", version)
	}

	major, minor, patch, preReleaseIDs := matches[1], matches[2], matches[3], matches[4]
	segments := []string{major, minor, patch}

	if preReleaseIDs == "" {
		segments = append(segments, maxValue)
	} else {
		ids := parseBySeparator(preReleaseIDs, anyPunctuationOrSeparator, toDecimalString)
		segments = append(segments, ids...)
	}

	return fromStringSlice(SemVer, version, segments)
}

func normalizeUnicode(s string) string {
	return norm.NFC.String(s)
}

// findNamedMatches returns a map of group names to matched strings from the
// leftmost match of the regular expression in version. A return value of nil
// indicates no match.
func findNamedMatches(version string, regex *regexp.Regexp) map[string]string {
	matches := regex.FindStringSubmatch(version)
	if matches == nil {
		return nil
	}

	groups := make(map[string]string, len(matches))
	for i, match := range matches {
		if match != "" {
			groups[regex.SubexpNames()[i]] = match
		}
	}
	return groups
}

// decimalStringConverter converts a string into a decimal number string. The
// input string is typically not expected to contain any numbers.
type decimalStringConverter func(string) string

func parseBySeparator(version string, separatorRegex *regexp.Regexp, convert decimalStringConverter) []string {
	parsed := []string{}
	for _, section := range separatorRegex.Split(version, -1) {
		section = wholeNumber.ReplaceAllString(section, delimitedSubsection)
		for _, piece := range strings.Split(section, delimiter) {
			parsed = maybeAppendDecimalString(parsed, piece, convert)
		}
	}
	return parsed
}

// maybeAppendDecimalString appends the string representation of a decimal
// number to the given string slice, if s is not the empty string. The convert
// converts a string to the proper decimal string form, which can be specific
// to the calling function.
func maybeAppendDecimalString(slice []string, s string, convert decimalStringConverter) []string {
	if s == "" {
		return slice
	}

	if !isNumber(s) {
		s = convert(s)
	}

	return append(slice, normalizeDecimal(s))
}

func isNumber(s string) bool {
	return decimalNumber.MatchString(s)
}

func normalizeDecimal(s string) string {
	// Any leading and trailing zeroes don't change the value of the decimal,
	// so strip them off to have a canonical string representation of the
	// decimal
	parts := strings.Split(s, ".")

	var normalized string
	if notZero.MatchString(parts[0]) {
		normalized = strings.TrimLeft(parts[0], "0")
	} else {
		normalized = "0"
	}

	if len(parts) == 2 && notZero.MatchString(parts[1]) {
		normalized = fmt.Sprintf("%s.%s", normalized, strings.TrimRight(parts[1], "0"))
	}

	return normalized
}

func toDecimalStringWithGenericPreReleaseIdentifierHandling(s string) string {
	if decimal, exists := genericPreReleaseIdentifiers[strings.ToLower(s)]; exists {
		return decimal
	}

	return toDecimalString(s)
}

func toDecimalString(s string) string {
	decimal := ""
	runeIndex := 0
	// The index returned when iterating over a string is the starting byte of
	// the current rune, which will jump by the number of bytes of the
	// previous rune. It is easier to keep track of the rune index if we do it
	// ourself.
	for _, r := range s {
		if runeIndex == 0 {
			decimal = fmt.Sprintf("%d", r)
			runeIndex++
			continue
		}

		if runeIndex == 1 {
			decimal += "."
		}

		// Pad to 10 digits using zeros because Unicode characters are 32-bit
		// integers and a 32-bit integer is a maximum of 10 digits long.
		decimal += fmt.Sprintf("%010d", r)
		runeIndex++
	}
	return decimal
}

func containsGenericPreReleaseIdentifierValue(numbers []string) bool {
	// Check if there is a negative number by checking for the minus sign.
	for _, n := range numbers {
		if n[0] == '-' {
			return true
		}
	}

	return false
}
