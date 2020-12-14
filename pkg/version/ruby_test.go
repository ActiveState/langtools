package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Many of the versions tested here are from
// https://github.com/rubygems/rubygems/blob/master/test/rubygems/test_gem_version.rb

var equalRubyVersions = [][]string{
	{
		"0.0.0", "000", "0", "", "   ", " ",
	},
	{
		"0.beta.1", "0.0.beta.1",
	},
	{
		"1", "000001", "1.0", "1.0.0", "1.0 ", " 1.0 ", "1.0\n", "\n1.0\n", "1.0",
	},
	{
		"1.0.0-1", "1-1",
	},
	{
		"1.2.b1", "1.2.b.1",
	},
	{
		"1.2.pre.1", "1.2.0.pre.1.0",
	},
	{
		"1.2", "1.2.0",
	},
	{
		"5.0.0.rc2", "5.0.rc2", "5.rc2",
	},
	{
		"5", "5.0.0",
	},
}

func TestParseRubyEqual(t *testing.T) {
	for _, versions := range equalRubyVersions {
		for i := 0; i < len(versions)-1; i++ {
			v1 := parseRubyOrFatal(t, versions[i])
			v2 := parseRubyOrFatal(t, versions[i+1])
			assert.True(
				t,
				Compare(v1, v2) == 0,
				"%v and %v should be equal", versions[i], versions[i+1],
			)
		}
	}
}

var invalidRubyVersions = []string{
	"whatever",
	"junk",
	"1.0\n2.0",
	"1..2",
	"1.ウ",
	"1.2 3.4",
	"2.3422222.222.222222222.22222.ads0as.dasd0.ddd2222.2.qd3e.",
}

func TestParseRubyInvalid(t *testing.T) {
	for _, invalidString := range invalidRubyVersions {
		v, err := ParseRuby(invalidString)
		assert.Nil(t, v)
		assert.Error(t, err, "%v should fail to parse", invalidString)
	}
}

var rubyTestStrings = []string{
	"0.0.beta",
	"0.beta.1",
	"0",
	"1.A",
	"1.0.a",
	"1-a",
	"1.0.0-alpha",
	"1.0.0-alpha.1",
	"1.0.0-beta.2",
	"1.0.0-beta.11",
	"1.0.0-rc.1",
	"1.0.0-1",
	"1",
	"1.1.rc10",
	"1.1",
	"1.2.0.a",
	"1.2.b1",
	"1.2.d.42",
	"1.2.pre.1",
	"1.2",
	"1.2.3.a.4",
	"1.2.3",
	"1.3",
	"1.8.2.A",
	"1.8.2.a",
	"1.8.2.a9",
	"1.8.2.a10",
	"1.8.2.b",
	"1.8.2",
	"1.9.a",
	"1.9.0.dev",
	"1.9.3.alpha.5",
	"1.9.3",
	"2.9.b",
	"2.9",
	"5.a",
	"5.0.0.rc2",
	"5.x",
	"5",
	"5.1",
	"5.2.4.a",
	"5.2.4.a10",
	"0005.2.4",
	"5.3",
	"6",
	"9.8.7",
	"9.8.8",
	"22.1.50.0.d",
	"22.1.50.0",
}

func TestParseRubyOrdering(t *testing.T) {
	for i := 0; i < len(rubyTestStrings)-1; i++ {
		v1 := parseRubyOrFatal(t, rubyTestStrings[i])
		v2 := parseRubyOrFatal(t, rubyTestStrings[i+1])
		assert.True(
			t,
			Compare(v1, v2) < 0,
			"%v should be less than %v", rubyTestStrings[i], rubyTestStrings[i+1],
		)
	}
}

func parseRubyOrFatal(t *testing.T, v string) *Version {
	ver, err := ParseRuby(v)
	require.NoError(t, err, "no error parsing %v as a ruby version", v)
	return ver
}
