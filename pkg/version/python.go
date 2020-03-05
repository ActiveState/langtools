package version

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ParsePython attempts to parse a version according to PEP440
// (https://www.python.org/dev/peps/pep-0440/) and falls back to legacy Python
// parsing if that fails.
func ParsePython(version string) (*Version, error) {
	result, err := parsePEP440(version)
	if err != nil {
		result, err = parseLegacyPython(version)
	}
	return result, err
}

const (
	// This regex was taken from PEP440 Appendix B for extracting the
	// components of a version identifier. It has been reformatted from the
	// original python to work for go.
	//
	// Note that the P<release> group allows for any number of '.' separated
	// segments. This causes a problem for us as we store versions into an
	// array, so if we don't have a limit on the number of segments in the
	// release we can end up comparing release version numbers against other
	// types of segments. To prevent this pep440MaxReleaseSegments is used to
	// ensure that we always compare the same type of segment data.
	pep440VersionPattern = `(?i)^\s*` +
		`v?` +
		`(?:` +
		`(?:(?P<epoch>[0-9]+)!)?` +
		`(?P<release>[0-9]+(?:\.[0-9]+)*)` +
		`(?P<pre>[-_\.]?(?P<pre_l>(a|b|c|rc|alpha|beta|pre|preview))[-_\.]?(?P<pre_n>[0-9]+)?)?` +
		`(?P<post>(?:-(?P<post_n1>[0-9]+))|(?:[-_\.]?(?P<post_l>post|rev|r)[-_\.]?(?P<post_n2>[0-9]+)?))?` +
		`(?P<dev>[-_\.]?(?P<dev_l>dev)[-_\.]?(?P<dev_n>[0-9]+)?)?` +
		`)` +
		`(?:\+(?P<local>[a-z0-9]+(?:[-_\.][a-z0-9]+)*))?` +
		`\s*$`

	// This is the number of indices in the final array that are reserved for
	// the release version. Changing this will cause comparison problems
	// between the old and new versions as the array indexes will no longer
	// match.
	pep440MaxReleaseSegments = 15

	// Values given to segment labels to ensure sort order is correct: dev,
	// pre-release, normal (implicit), post-release
	pep440DevRelease   = "-4"
	pep440AlphaRelease = "-3"
	pep440BetaRelease  = "-2"
	pep440RCRelease    = "-1"
	pep440Implicit     = "0"
	pep440PostRelease  = "1"
)

var pep440NormalizationRegex = regexp.MustCompile(pep440VersionPattern)

// parsePEP440 parses version using the version parsing algorithm defined in
// python PEP 440 (https://www.python.org/dev/peps/pep-0440/).  Normalization,
// as defined in PEP 440, is performed on version before parsing occurs. If
// version is a local version identifier its local segment will be part of the
// result.
func parsePEP440(version string) (*Version, error) {
	matches := findNamedMatches(version, pep440NormalizationRegex)
	if matches == nil {
		return nil, fmt.Errorf("not PEP440 version: %s", version)
	}

	releaseSegments := strings.Split(matches["release"], ".")
	if len(releaseSegments) > pep440MaxReleaseSegments {
		return nil, fmt.Errorf("exceeds max number of release segments: %s", version)
	}

	for i := len(releaseSegments); i < pep440MaxReleaseSegments; i++ {
		releaseSegments = append(releaseSegments, pep440Implicit)
	}

	preLabel, preNumber := pep440PreReleaseSegments(matches)
	postLabel, postNumber := pep440PostReleaseSegments(matches)
	devLabel, devNumber := pep440DevReleaseSegments(matches)

	// The general sort order is: dev, pre, <nothing>, post, local
	// Ex. 1.0.dev1 < 1.0a1.dev1 < 1.0.a1 < 1.0 < 1.0.post1.dev1 < 1.0.post1
	// The only case our normal sorting doesn't handle is the first one,
	// making a dev release before a pre-release. To make that happen we abuse
	// the preLabel by assigning it a value that will sort before any normal
	// preLabel value.
	//
	// See https://github.com/pypa/packaging/pull/1/files
	// packaging/version.py:286 (_cmpkey()) for the python implementation of
	// PEP440 that this was derived from.
	if preLabel == pep440Implicit && postLabel == pep440Implicit && devLabel == pep440DevRelease {
		preLabel = pep440DevRelease
	}

	segments := make([]string, 0, 12)
	segments = append(segments, pep440EpochSegment(matches))
	segments = append(segments, releaseSegments...)
	segments = append(segments,
		preLabel, preNumber,
		postLabel, postNumber,
		devLabel, devNumber,
	)
	segments = append(segments, pep440LocalSegments(matches)...)

	return fromStringSlice(PythonPEP440, version, segments)
}

func pep440EpochSegment(matches map[string]string) string {
	if v, ok := matches["epoch"]; ok {
		return v
	}
	return pep440Implicit
}

func pep440PreReleaseSegments(matches map[string]string) (string, string) {
	if _, ok := matches["pre"]; !ok {
		return pep440Implicit, pep440Implicit
	}

	var label string
	switch strings.ToLower(matches["pre_l"]) {
	case "a", "alpha":
		label = pep440AlphaRelease
	case "b", "beta":
		label = pep440BetaRelease
	case "c", "rc", "pre", "preview":
		label = pep440RCRelease
	default:
		panic("PEP440 regex has bad pre-release label match group")
	}

	if n, ok := matches["pre_n"]; ok {
		return label, n
	}

	return label, pep440Implicit
}

func pep440PostReleaseSegments(matches map[string]string) (string, string) {
	if _, ok := matches["post"]; !ok {
		return pep440Implicit, pep440Implicit
	}

	if n, ok := matches["post_n1"]; ok {
		return pep440PostRelease, n
	}

	if n, ok := matches["post_n2"]; ok {
		return pep440PostRelease, n
	}

	return pep440PostRelease, pep440Implicit
}

func pep440DevReleaseSegments(matches map[string]string) (string, string) {
	if _, ok := matches["dev"]; !ok {
		return pep440Implicit, pep440Implicit
	}

	if n, ok := matches["dev_n"]; ok {
		return pep440DevRelease, n
	}

	return pep440DevRelease, pep440Implicit
}

func pep440LocalSegments(matches map[string]string) []string {
	local, ok := matches["local"]
	if !ok {
		return nil
	}

	// "With a local version, in addition to the use of . as a separator of
	// segments, the use of - and _ is also acceptable." - PEP440
	local = strings.ReplaceAll(local, "-", ".")
	local = strings.ReplaceAll(local, "_", ".")

	var segments []string
	for _, s := range strings.Split(local, ".") {
		// Local strings are compared with case insensitivity
		s = strings.ToLower(s)

		// Numeric segments are supposed to always compare greater than
		// lexicographic segments. Because local lexicographic segments may
		// only be ASCII, prepending 128 works.
		if _, err := strconv.Atoi(s); err == nil {
			segments = append(segments, "128")
			segments = append(segments, s)
		} else {
			segments = append(segments, toDecimalString(s))
		}
	}

	return segments
}

var legacyPythonSegmentsRegex = regexp.MustCompile(`\d+|[a-z]+|\.|-`)

var legacyPythonReplacements = map[string]string{
	"pre":     "c",
	"preview": "c",
	"-":       "final-",
	"rc":      "c",
	"dev":     "@",
}

func splitLegacyPythonSegments(version string) []string {
	// Split the version based on matches in legacyPythonSegmentsRegex, but
	// keep both the matches and the things between the matches
	b := []byte(version)
	repl := func(in []byte) []byte {
		out := make([]byte, len(in))
		copy(out, in)
		out = append(out, '\x00')
		return out
	}
	b = legacyPythonSegmentsRegex.ReplaceAllFunc(b, repl)
	bSegments := bytes.Split(b, []byte{'\x00'})

	var segments []string
	for _, bSegment := range bSegments {
		segment := string(bSegment)

		if replacement, ok := legacyPythonReplacements[segment]; ok {
			segment = replacement
		}

		if segment == "" || segment == "." {
			continue
		}

		if numSegment, err := strconv.Atoi(segment); err == nil {
			if len(segment) <= 8 {
				segment = fmt.Sprintf("%08d", numSegment)
			}
		} else {
			segment = "*" + segment
		}

		segments = append(segments, segment)
	}

	segments = append(segments, "*final")

	return segments
}

// parseLegacyPython parses as described at
// https://github.com/pypa/packaging/blob/19.2/packaging/version.py#L124-L176
//
// A legacy Python version will always start with -1 in order to sort as
// before all PEP440 versions.
func parseLegacyPython(version string) (*Version, error) {
	segments := []string{}
	for _, segment := range splitLegacyPythonSegments(strings.ToLower(version)) {
		if strings.HasPrefix(segment, "*") {
			if segment < "*final" {
				for len(segments) > 0 && segments[len(segments)-1] == "*final-" {
					segments = segments[:len(segments)-1]
				}
			}

			// Remove trailing zeros from each series of numeric segments
			for len(segments) > 0 && segments[len(segments)-1] == "00000000" {
				segments = segments[:len(segments)-1]
			}
		}

		segments = append(segments, segment)
	}

	// Legacy versions are always compared lexicographically
	for i, segment := range segments {
		segments[i] = toDecimalString(segment)
	}

	// Epoch of -1 makes all legacy versions come before all PEP440 versions.
	epoch := "-1"
	segments = append([]string{epoch}, segments...)

	return fromStringSlice(PythonLegacy, version, segments)
}
