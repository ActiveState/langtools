package version

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	phpAliasRegex = regexp.MustCompile(
		`^([^,\s]+) +as +([^,\s]+)$`,
	)
	phpAtStabilitiesRegex = regexp.MustCompile(
		`(?i)@(?:stable|RC|beta|alpha|dev)$`,
	)
	phpBuildRegex = regexp.MustCompile(
		`^([^,\s+]+)\+[^\s]+$`,
	)
	phpClassicalRegex = regexp.MustCompile(
		`(?i)^v?(\d{1,5})(\.\d+)?(\.\d+)?(\.\d+)?[._-]?(?:(stable|beta|b|RC|alpha|a|patch|pl|p)((?:[.-]?\d+)*)?)?([.-]?dev)?$`,
	)
	phpDatetimeRegex = regexp.MustCompile(
		`(?i)^v?(\d{4}(?:[.:-]?\d{2}){1,6}(?:[.:-]?\d{1,3})?)[._-]?(?:(stable|beta|b|RC|alpha|a|patch|pl|p)((?:[.-]?\d+)*)?)?([.-]?dev)?$`,
	)
	phpNondigitRegex = regexp.MustCompile(
		`\D`,
	)
	phpDigitWordRegex = regexp.MustCompile(
		`(\d)([a-zA-Z])`,
	)
	phpWordDigitRegex = regexp.MustCompile(
		`([a-zA-Z])(\d)`,
	)
)

// ParsePHP attempts to parse a version according to the same rules used by
// composer (https://github.com/composer/semver)
func ParsePHP(version string) (*Version, error) {
	original := version

	version, err := normalizePHP(version)
	if err != nil {
		return nil, err
	}

	version = strings.ReplaceAll(version, "_", ".")
	version = strings.ReplaceAll(version, "-", ".")
	version = strings.ReplaceAll(version, "+", ".")

	version = phpDigitWordRegex.ReplaceAllString(version, "$1.$2")
	version = phpWordDigitRegex.ReplaceAllString(version, "$1.$2")

	segments := strings.Split(version, ".")
	numericSegments := convertPHPSegments(segments)
	return fromStringSlice(PHP, original, numericSegments)
}

func convertPHPSegments(segments []string) []string {
	results := []string{}
	leadingSegmentCount := 0
	hasSpecial := false
	lastIsSpecial := false
	for _, segment := range segments {
		switch segment {
		case "dev":
			segment = "-4"
			hasSpecial = true
			lastIsSpecial = true
		case "alpha":
			segment = "-3"
			hasSpecial = true
			lastIsSpecial = true
		case "beta":
			segment = "-2"
			hasSpecial = true
			lastIsSpecial = true
		case "RC":
			segment = "-1"
			hasSpecial = true
			lastIsSpecial = true
		case "patch":
			segment = "0.5"
			hasSpecial = true
			lastIsSpecial = true
		default:
			if !hasSpecial {
				leadingSegmentCount++
			}
			lastIsSpecial = false
		}
		results = append(results, segment)
	}

	// Special asinine "datetime" version handling. This is probably a bug
	// in the semver PHP library that we are doing our best to reproduce...
	if leadingSegmentCount < 4 {
		var value string
		if len(results) > leadingSegmentCount && results[leadingSegmentCount] == "0.5" {
			value = "1000000000"
		} else {
			value = "-0.5"
		}
		results = append(
			results[:leadingSegmentCount],
			append([]string{value}, results[leadingSegmentCount:]...)...,
		)
	}

	// Ensure that "1.0.patch" < "1.0.patch.0".
	if lastIsSpecial {
		results = append(results, "-0.5")
	}

	return results
}

func expandPHPStability(stability string) string {
	switch strings.ToLower(stability) {
	case "a":
		return "alpha"
	case "b":
		return "beta"
	case "p":
		return "patch"
	case "pl":
		return "patch"
	case "rc":
		return "RC"
	default:
		return stability
	}
}

func normalizePHP(version string) (string, error) {
	original := version

	// Extra whitespace is tolerated
	version = strings.TrimSpace(version)

	// Case doesn't matter
	version = strings.ToLower(version)

	// Remove aliasing
	matches := phpAliasRegex.FindStringSubmatch(version)
	if len(matches) > 1 {
		version = matches[1]
	}

	// Remove stability suffix
	loc := phpAtStabilitiesRegex.FindStringIndex(version)
	if loc != nil {
		version = version[:loc[0]]
	}

	// Remove build metadata
	matches = phpBuildRegex.FindStringSubmatch(version)
	if len(matches) > 1 {
		version = matches[1]
	}

	// Try normal matching first
	index := 0
	matches = phpClassicalRegex.FindStringSubmatch(version)
	if len(matches) > 4 {
		if matches[2] == "" {
			matches[2] = ".0"
		}
		if matches[3] == "" {
			matches[3] = ".0"
		}
		if matches[4] == "" {
			matches[4] = ".0"
		}
		version = matches[1] + matches[2] + matches[3] + matches[4]
		index = 5
	}
	if len(matches) == 0 {
		// Then try datetime matching
		matches = phpDatetimeRegex.FindStringSubmatch(version)
		if len(matches) > 1 {
			version = phpNondigitRegex.ReplaceAllLiteralString(matches[1], ".")
			index = 2
		}
	}

	// Add version modifiers
	if index != 0 {
		if matches[index] != "" {
			if matches[index] == "stable" {
				return version, nil
			}
			version = version + "-"
			version = version + expandPHPStability(matches[index])
			if matches[index+1] != "" {
				version = version + strings.TrimLeft(matches[index+1], ".-")
			}
		}

		if matches[index+2] != "" {
			version = version + "-dev"
		}

		return version, nil
	}

	return "", fmt.Errorf("invalid php version: %v", original)
}
