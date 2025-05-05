package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var equalGoVersions = [][]string{
	{
		"v0.0.0", "v000", "v0",
	},
	{
		"v1", "v000001", "v1.0", "v1.0.0", "v1.0 ",
	},
	{
		"v1.2.b1", "v1.2.b.1",
	},
	{
		"v1.2", "v1.2.0",
	},
	{
		"v5", "v5.0.0",
	},
}

func TestParseGoEqual(t *testing.T) {
	for _, versions := range equalGoVersions {
		for i := 0; i < len(versions)-1; i++ {
			v1 := parseGoOrFatal(t, versions[i])
			v2 := parseGoOrFatal(t, versions[i+1])
			assert.True(
				t,
				Compare(v1, v2) == 0,
				"%v and %v should be equal, but '%v' != '%v'", versions[i], versions[i+1], v1, v2,
			)
		}
	}
}

var invalidGoVersions = []string{
	"whatever",
	"junk",
	"1.0\n2.0",
	"1..2",
	"1.ウ",
	"1.2 3.4",
	"2.3422222.222.222222222.22222.ads0as.dasd0.ddd2222.2.qd3e.",
}

func TestParseGoInvalid(t *testing.T) {
	for _, invalidString := range invalidGoVersions {
		v, err := ParseGo(invalidString)
		assert.Nil(t, v)
		assert.Error(t, err, "%v should fail to parse", invalidString)
	}
}

var goTestStrings = []string{
	"v0.beta.1",
	"v0",
	"v1",
	"v1.1",
	"v1.2",
	"v1.2.3",
	"v1.3",
	"v1.8.2",
	"v1.9.3",
	"v2.9",
	"v5",
	"v5.1",
	"v5.3",
	"v6",
	"v9.8.7",
	"v9.8.8",
	"v22.1.50.0",
}

func TestParseGoOrdering(t *testing.T) {
	for i := 0; i < len(goTestStrings)-1; i++ {
		v1 := parseGoOrFatal(t, goTestStrings[i])
		v2 := parseGoOrFatal(t, goTestStrings[i+1])
		assert.True(
			t,
			Compare(v1, v2) < 0,
			"%v should be less than %v", goTestStrings[i], goTestStrings[i+1],
		)
	}
}

func TestParseGo(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected []string
	}{
		{"Numbers", "v0", []string{"0"}},
		{"Numbers", "v1", []string{"1"}},
		{"Numbers", "v1.0", []string{"1"}},
		{"Numbers", "v0.92", []string{"0", "92"}},
		{"Numbers", "v1-1.2", []string{"1", "1", "2"}},
		{"Sequential Dots", "v1..2", []string{"1", "2"}},
		{"Sequential Dashes", "v1--2", []string{"1", "2"}},
		{"Sequential Dot Dash", "v1.-2", []string{"1", "2"}},
		{"Uppercase A", "vA1", []string{"65", "1"}},
		{"Lowercase a", "va1", []string{"97", "1"}},
		{"Single Unicode", "v小1", []string{"23567", "1"}},
		{"Ascii Word", "v1.0bet", []string{"1", "0", "98.00000001010000000116"}},
		{"Unicode Word", "v小寸-1.1", []string{"23567.0000023544", "1", "1"}},
		{"Unicode Separators", "v1 2\u20013\u2002\u20034", []string{"1", "2", "3", "4"}},
		{"Normalizes Unicode", "ve\u0301", []string{"233"}},
		{
			"Splits On Space",
			"v10 Generic 142910-17",
			[]string{
				"10",
				"71.000000010100000001100000000101000000011400000001050000000099",
				"142910",
				"17",
			},
		},
		{"Drops Leading Zeros", "v100.02.01", []string{"100", "2", "1"}},
		{"Pre-Release Identifier", "v1.0-alpha", []string{"1", "0", "-26"}},
		{"Pre-Release Identifier Ignores Case", "v1.0-AlPHa", []string{"1", "0", "-26"}},
		{"Pre-Release Identifier In Middle", "v1.0-alpha.1", []string{"1", "0", "-26", "1"}},
		{"2 Pre-Release Identifiers", "v1.0-alpha.beta", []string{"1", "0", "-26", "-25"}},
		{"Pre-Release Identifier Beta", "v1.0-beta", []string{"1", "0", "-25"}},
		{"Pre-Release Identifier RC", "v1.0-rc", []string{"1", "0", "-1"}},
		{"Timestamp-commits", "v1.2.3-20191109021931-daa7c04131f5", []string{"1", "2", "3", "20191109021931"}},
		{"Timestamp-dot-commits", "v1.23.456.789.20191109021931-caa7c04131f6", []string{"1", "23", "456", "789", "20191109021931"}},
		{"Timestamp-svn-commits", "v1.23.456.789-20191109021931-000000001234", []string{"1", "23", "456", "789", "20191109021931"}},
		{"Timestamp-svn-dot-commits", "v9.87.654.321.20191109021931-000000001234", []string{"9", "87", "654", "321", "20191109021931"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := parseGoOrFatal(t, tt.version)
			assert.Equal(t, Go, actual.ParsedAs, "got expected ParsedAs value")
			assertDecimalEqualString(t, tt.expected, actual.Decimal)
			assertDecimalEqualDecimal(t, tt.expected, actual.Decimal)
		})
	}
}


func parseGoOrFatal(t *testing.T, v string) *Version {
	ver, err := ParseGo(v)
	require.NoError(t, err, "no error parsing %v as a go version", v)
	return ver
}
