package version

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	rubyVersionPattern = `\A\s*[0-9]+(\.[0-9a-zA-Z]+)*(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?\s*\z`
	rubySegmentPattern = `[0-9]+|[A-Za-z]+`
)

var (
	rubyVersionRegex = regexp.MustCompile(rubyVersionPattern)
	rubySegmentRegex = regexp.MustCompile(rubySegmentPattern)
)

// ParseRuby attempts to parse a version according to the same rules used by
// rubygems (https://github.com/rubygems/rubygems)
func ParseRuby(version string) (*Version, error) {
	v := strings.TrimSpace(version)
	if v == "" {
		v = "0"
	}

	if !rubyVersionRegex.MatchString(v) {
		return nil, fmt.Errorf("invalid ruby version: %v", version)
	}

	v = strings.ReplaceAll(v, "-", ".pre.")

	segments := splitSegments(v)
	if len(segments) == 0 {
		segments = []string{"0"}
	}

	output := []string{}
	for _, segment := range segments {
		_, err := strconv.Atoi(segment)
		if err != nil {
			// A string segment must compare less than any numeric segment
			output = append(output, "-1")
			output = append(output, asciiToDecimalString(segment))
		} else {
			output = append(output, segment)
		}
	}

	return fromStringSlice(Ruby, version, output)
}

func splitSegments(version string) []string {
	segments := rubySegmentRegex.FindAllString(version, -1)

	// Create two segment groups by splitting at the first non-integer
	// Also normalize integer formats as we go (e.g. change "002" to "2")
	before := []string{}
	after := []string{}
	i := 0
	for i < len(segments) {
		s, err := strconv.Atoi(segments[i])
		if err != nil {
			break
		}

		before = append(before, strconv.Itoa(s))
		i++
	}
	for i < len(segments) {
		s, err := strconv.Atoi(segments[i])
		if err != nil {
			after = append(after, segments[i])
		} else {
			after = append(after, strconv.Itoa(s))
		}
		i++
	}

	before = dropTrailingZeroes(before)
	after = dropTrailingZeroes(after)

	return append(before, after...)
}

func dropTrailingZeroes(segments []string) []string {
	lastNonzeroIndex := len(segments) - 1
	for i := lastNonzeroIndex; i >= 0; i-- {
		if segments[i] != "0" {
			break
		}
		lastNonzeroIndex--
	}
	return segments[0 : lastNonzeroIndex+1]
}
