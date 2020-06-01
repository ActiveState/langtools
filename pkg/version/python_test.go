package version

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePython(t *testing.T) {
	tests := map[ParsedAs]map[string]struct {
		version  string
		expected []string
	}{
		PythonPEP440: {
			"Minimal": {
				version: "1",
				expected: []string{
					"0",                                                                       // epoch
					"1", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", // release
					"0", "0", // pre-release
					"0", "0", // post-release
					"0", "0", // dev release
				},
			},
			"Leading v is ignored": {
				version: "v1",
				expected: []string{
					"0",
					"1", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0",
					"0", "0",
					"0", "0",
					"0", "0",
				},
			},
			"Maximum release digits used": {
				version: "1.2.3.4.5.6.7.8.9.10.11.12.13.14.15",
				expected: []string{
					"0",
					"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15",
					"0", "0",
					"0", "0",
					"0", "0",
				},
			},
			"Alpha": {
				version: "1a2",
				expected: []string{
					"0",
					"1", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0",
					"-3", "2",
					"0", "0",
					"0", "0",
				},
			},
			"Beta": {
				version: "1b2",
				expected: []string{
					"0",
					"1", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0",
					"-2", "2",
					"0", "0",
					"0", "0",
				},
			},
			"RC": {
				version: "1rc2",
				expected: []string{
					"0",
					"1", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0",
					"-1", "2",
					"0", "0",
					"0", "0",
				},
			},
			"C is RC": {
				version: "1c2",
				expected: []string{
					"0",
					"1", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0",
					"-1", "2",
					"0", "0",
					"0", "0",
				},
			},
			"Canonical Public Version Identifier": {
				version: "99!1.2.3.4.5a6.post7.dev8",
				expected: []string{
					"99",
					"1", "2", "3", "4", "5", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0",
					"-3", "6",
					"1", "7",
					"-4", "8",
				},
			},
			"Local Version Identifier": {
				version: "1+aA.2B.3",
				expected: []string{
					"0",
					"1", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0", "0",
					"0", "0",
					"0", "0",
					"0", "0",
					"97.0000000097", "50.0000000098", "128", "3",
				},
			},
		},
		PythonLegacy: {
			"Fall back to legacy version parsing": {
				version: "2.6.0-0.1",
				expected: []string{
					"-1", // epoch is always -1 for legacy
					"48.0000000048000000004800000000480000000048000000004800000000480000000050", // "00000002"
					"48.0000000048000000004800000000480000000048000000004800000000480000000054", // "00000006"
					"42.000000010200000001050000000110000000009700000001080000000045",           // "*final-"
					"48.0000000048000000004800000000480000000048000000004800000000480000000048", // "00000000"
					"48.0000000048000000004800000000480000000048000000004800000000480000000049", // "00000001"
					"42.00000001020000000105000000011000000000970000000108",                     // "*final"
				},
			},
		},
	}

	for pa, cases := range tests {
		for name, tt := range cases {
			t.Run(name, func(t *testing.T) {
				actual, err := ParsePython(tt.version)
				require.NoError(t, err)
				assert.Equal(t, pa, actual.ParsedAs, "got expected ParsedAs value")
				assertDecimalEqualString(t, tt.expected, actual.Decimal)
				assertDecimalEqualDecimal(t, tt.expected, actual.Decimal)
			})
		}
	}
}

// Many of these tests are from
// https://github.com/pypa/packaging/blob/19.2/tests/test_version.py
//
// They can be verified via https://pypi.org/project/packaging/19.2/
// as follows:
//
// $ python3
// >>> from packaging import version
// >>> version.parse("some version") < version.parse("another version")
var pythonTestStrings = []string{
	// Legacy version tests, implicit epoch of -1
	"  hmm",
	"a cat is fine too",
	"a",
	"b",
	"foobar",
	"lolwut",
	"0000000011g",
	"1.13++",
	"000000011g",
	"2.0b1pl0",
	"2e6",
	"2g6",
	"2.6.0-0.1pre6",
	"2.6.0-0.1-pre7",
	"2.6.0-0.1",
	"2.6.0-0.2",
	"2.6.0-0.92",
	"2.7.0-0.92",
	"2.16.0-0.92",
	"3.2pl0",
	"3.4j",
	"5.5.kw",
	"11g",
	"012g",

	// Implicit epoch of 0
	"1.0.dev0",
	"1.0.dev456",
	"1.0a0",
	"1.0a1",
	"1.0a2.dev456",
	"1.0a12.dev456",
	"1.0a12",
	"1.0b1.dev456",
	"1.0b2",
	"1.0b2.post345.dev456",
	"1.0b2.post345",
	"1.0b2-346",
	"1.0rc1.dev456",
	"1.0rc1",
	"1.0rc2",
	"1.0c3",
	"1.0",
	"1.0+abc.5",
	"1.0+abc.7",
	"1.0+5",
	"1.0.post456.dev34",
	"1.0.post456",
	"1.0.1.2.3.4.5.6.7.8.9.1.2.3.4",
	"1.1.dev1",
	"1.2",
	"1.2+123abc",
	"1.2+123abc456",
	"1.2+abc",
	"1.2+abc123",
	"1.2+abc123def",
	"1.2+abcd",
	"1.2+def",
	"1.2+1",
	"1.2+05",
	"1.2+12",
	"1.2+25",
	"1.2+123",
	"1.2+123.abc",
	"1.2+123-def",
	"1.2+123_gg",
	"1.2+0124",
	"1.2+1234.abc",
	"1.2+123456",
	"1.2.r32+123456",
	"1.2.rev33+123456",

	// Explicit epoch of 1
	"1!1.0.dev456",
	"1!1.0a1",
	"1!1.0a2.dev456",
	"1!1.0a12.dev456",
	"1!1.0a12",
	"1!1.0b1.dev456",
	"1!1.0b2",
	"1!1.0b2.post345.dev456",
	"1!1.0b2.post345",
	"1!1.0b2-346",
	"1!1.0c1.dev456",
	"1!1.0c1",
	"1!1.0rc2",
	"1!1.0c3",
	"1!1.0",
	"1!1.0.post456.dev34",
	"1!1.0.post456",
	"1!1.1.dev1",
	"1!1.2+123abc",
	"1!1.2+123abc456",
	"1!1.2+abc",
	"1!1.2+abc123",
	"1!1.2+abc123def",
	"1!1.2+1234.abc",
	"1!1.2+123456",
	"1!1.2.r32+123456",
	"1!1.2.rev33+123456",
}

func TestParsePythonOrdering(t *testing.T) {
	for i := 0; i < len(pythonTestStrings)-1; i++ {
		v1 := parsePythonOrFatal(t, pythonTestStrings[i])
		v2 := parsePythonOrFatal(t, pythonTestStrings[i+1])
		assert.True(
			t,
			Compare(v1, v2) < 0,
			fmt.Sprintf("%s < %s", pythonTestStrings[i], pythonTestStrings[i+1]),
		)
	}
}

func parsePythonOrFatal(t *testing.T, v string) *Version {
	ver, err := ParsePython(v)
	assert.NoError(t, err, "no error parsing %s as a python version", v)
	return ver
}
