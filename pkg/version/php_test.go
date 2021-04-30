package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Many of the test inputs here are from
// https://github.com/composer/semver/tree/main/tests

var normalizePHPTests = [][]string{
	{" 1.0.0", "1.0.0.0"},
	{"0", "0.0.0.0"},
	{"0.000.103.204", "0.000.103.204"},
	{"00.01.03.04", "00.01.03.04"},
	{"000.001.003.004", "000.001.003.004"},
	{"0000000", "0000000"},
	{"0000000000001", "0000000000001"},
	{"0700", "0700.0.0.0"},
	{"1.0", "1.0.0.0"},
	{"1.0-dev", "1.0.0.0-dev"},
	{"1.0.0 ", "1.0.0.0"},
	{"1.0.0", "1.0.0.0"},
	{"1.0.0+foo as 2.0", "1.0.0.0"},
	{"1.0.0+foo", "1.0.0.0"},
	{"1.0.0+foo@dev", "1.0.0.0"},
	{"1.0.0-alpha-2.1-3+foo", "1.0.0.0-alpha2.1-3"},
	{"1.0.0-alpha.3.1+foo", "1.0.0.0-alpha3.1"},
	{"1.0.0-alpha2.1+foo", "1.0.0.0-alpha2.1"},
	{"1.0.0-beta.5+foo", "1.0.0.0-beta5"},
	{"1.0.0-rC15-dev", "1.0.0.0-RC15-dev"},
	{"1.0.0-rc1", "1.0.0.0-RC1"},
	{"1.0.0-stable", "1.0.0.0"},
	{"1.0.0.RC.15-dev", "1.0.0.0-RC15-dev"},
	{"1.0.0.pl3-dev", "1.0.0.0-patch3-dev"},
	{"1.0.0RC1dev", "1.0.0.0-RC1-dev"},
	{"1.13.11.0-beta0", "1.13.11.0-beta0"},
	{"1.2.3.4", "1.2.3.4"},
	{"10.4.13-b", "10.4.13.0-beta"},
	{"10.4.13-b5", "10.4.13.0-beta5"},
	{"10.4.13-beta", "10.4.13.0-beta"},
	{"10.4.13beta.2", "10.4.13.0-beta2"},
	{"10.4.13beta2", "10.4.13.0-beta2"},
	{"2010-01-02", "2010.01.02"},
	{"2010-01-02.5", "2010.01.02.5"},
	{"2010.01", "2010.01.0.0"},
	{"2010.01.02", "2010.01.02.0"},
	{"2010.1.555", "2010.1.555.0"},
	{"2010.1.555", "2010.1.555.0"},
	{"2010.10.200", "2010.10.200.0"},
	{"20100102-203040", "20100102.203040"},
	{"20100102-203040-p1", "20100102.203040-patch1"},
	{"20100102203040-10", "20100102203040.10"},
	{"2012.06.07", "2012.06.07.0"},
	{"201903.0", "201903.0"},
	{"201903.0-p2", "201903.0-patch2"},
	{"v1.0.0", "1.0.0.0"},
	{"v1.13.11-beta.0", "1.13.11.0-beta0"},
	{"v20100102", "20100102"},
}

func TestNormalizePHP(t *testing.T) {
	for _, test := range normalizePHPTests {
		input := test[0]
		expected := test[1]
		output, err := normalizePHP(input)
		assert.NoError(t, err)
		assert.Equal(t, expected, output)
	}
}

var invalidPHPVersions = []string{
	" as ",
	" as 1.2",
	"",
	"1.*",
	"1.0 .2",
	"1.0.0#",
	"1.0.0+foo bar",
	"1.0.0-dev<1.0.5-dev",
	"1.0.0-meh",
	"1.0.0.0.0",
	"1.0.0.abc",
	"1.0.0.alpha2.99.beta",
	"1.0.0<1.0.5-dev",
	"1.0.1-SNAPSHOT",
	"1.0.alpha.beta",
	"1.p.0.p",
	"1.x",
	"2010-1-555",
	"20100102.203040.0.1",
	"2147483647.0.0.0",
	"^",
	"^1",
	"^8 || ^",
	"a",
	"alpha",
	"feature-foo",
	"foo bar-dev",
	"~",
	"~1 ~",
	"~1",

	// These may be allowed as "versions" in certain PHP scenarios, but we
	// don't allow them because they are not sortable
	"041.x-dev",
	"1.x-dev",
	"2.0.*-dev",
	"20100102.203040.x-dev",
	"20100102.x-dev",
	"2010102.203040dev",
	"201903.x-dev",
	"DEV-FOOBAR",
	"dev-041.003",
	"dev-1.0.0-dev<1.0.5-dev",
	"dev-feature+issue-1",
	"dev-feature-foo",
	"dev-feature/foo",
	"dev-foo bar",
	"dev-load-varnish-only-when-used as ^2.0",
	"dev-load-varnish-only-when-used@dev as ^2.0@dev",
	"dev-load-varnish-only-when-used@stable",
	"dev-master as 1.0.0",
	"dev-master",
	"dev-trunk",
	"master",
}

func TestInvalidPHPVersions(t *testing.T) {
	for _, test := range invalidPHPVersions {
		v, err := ParsePHP(test)
		assert.Nil(t, v)
		assert.Error(t, err, "%v should fail to parse", test)
	}
}

var testParsePHPEqualInputs = [][]string{
	{"0", "0.0", "0.0.0", "0000", "0.0.0.0-stable"},
	{"000000", "0000000", "00000000"},
	{"1a", "1alpha"},
	{"2.b", "2-beta", "2-b", "2.beta"},
	{"3RC", "3.0.0.0-rc"},
	{"4dev", "4.dev", "4-dev"},
	{"5.2.p", "5.2.0-patch", "5.2.0.0pl"},
	{"6p0", "6.p-0", "6.0.0.0.patch.0"},
	{"7010-01-02", "7010-01-02."},
	{"8010000102.", "8010000102"},
}

func TestParsePHPEqual(t *testing.T) {
	for _, versions := range testParsePHPEqualInputs {
		for i := 0; i < len(versions)-1; i++ {
			v1 := parsePHPOrFatal(t, versions[i])
			v2 := parsePHPOrFatal(t, versions[i+1])
			assert.True(
				t,
				Compare(v1, v2) == 0,
				"%v and %v should be equal", versions[i], versions[i+1],
			)
		}
	}
}

var testParsePHPOrderInputs = []string{
	"0000000",
	"0",
	"0000000000001",
	"1.0.0.dev",
	"1.0.0.alpha",
	"1.0.0.alpha00000000000",
	"1.0.0.alpha1",
	"1.0.0.alpha2.99.1",
	"1.0.0.beta",
	"1.0.0.beta0.09",
	"1.0.0.beta009",
	"1.0.0.RC",
	"1.0.0",
	"1.0.0.p",
	"1.0.0.patch0",
	"1.0.0.patch1.0",
	"1.0.0.patch2",
	"1.0.0.1",
	"1.2.3",
	"1.2.3.4",
	"2.0.0.RC",
	"2.0.0-stable",
	"2.0.0.pl",
	"2.1",
	"2.2",
	"2.2.p",
	"2.2.0.1",
	"4.3.0",
	"5.3.dev",
	"5.3.0",
	"5.4",
	"5.9999999",
	"5.9999999.9999999",
	"5.9999999.9999999.p",
	"5.9999999.9999999.9999999",
	"5.9999999.9999999.9999999.p",
	"5.10000000",
	"5.10000001",
	"6.0",
	"2010-01-02-dev",
	"2010-01-02-a",
	"2010-01-02",
	"2010.01.02.dev",
	"2010.01.02.a",
	"2010.01.02-stable",
	"2010.01.02.p",
	"2010.01.02.p0",
	"2010.01.02.p1",
	"2010-01-02-p",
	"2010-01-02-p0",
	"2010-01-02-p1",
	"2010.1.555",
	"2010.10.200",
	"2010.11",
	"20112.dev",
	"20112.0alpha",
	"20112.beta",
	"20112.",
	"20112.0p",
	"20112.10.10.10",
	"20112.203040dev",
	"20112.203040alpha",
	"20112.203040.0beta",
	"20112.203040",
	"20112.203040.p1",
	"20112.203040.0p0123",
	"20113",
	"201101",
	"201102.dev",
	"201102.alpha",
	"201102.beta",
	"201102.",
	"201102.0alpha",
	"201102.0p",
	"201102.10.10.10",
	"201102.203040dev",
	"201102.203040alpha",
	"201102.203040",
	"201102.203040.0beta",
	"201102.203040.0",
	"201102.203040.0p0123",
	"201102-203040-p",
	"201102-203040-p1",
	"201102-p",
	"201103",
	"2010101",
	"2010102.dev",
	"2010102.beta",
	"2010102.",
	"2010102-p",
	"20100101",
	"20100102.dev",
	"20100102.alpha",
	"20100102.beta",
	"20100102.",
	"20100102.0alpha",
	"20100102.0p",
	"20100102.10.10.10",
	"20100102.203040dev",
	"20100102.203040alpha",
	"20100102.203040",
	"20100102.203040.0beta",
	"20100102.203040.0",
	"20100102.203040.0p0123",
	"20100102-203040-p",
	"20100102-203040-p1",
	"20100102-p",
	"20100103",
	"201000101",
	"201000102.dev",
	"201000102.alpha",
	"201000102.beta",
	"201000102.",
	"201000102-p",
	"201000103",
	"2010000101",
	"2010000102.dev",
	"2010000102.alpha",
	"2010000102.beta",
	"2010000102.",
	"2010000102.0alpha",
	"2010000102.0p",
	"2010000102.10.10.10",
	"2010000102.203040dev",
	"2010000102.203040alpha",
	"2010000102.203040",
	"2010000102.203040.0beta",
	"2010000102.203040.0",
	"2010000102.203040.0p0123",
	"2010000102-203040-p",
	"2010000102-203040-p1",
	"2010000102-999999999-p1",
	"2010000102-p",
	"2010000103",
}

func TestParsePHPOrdering(t *testing.T) {
	for i := 0; i < len(testParsePHPOrderInputs)-1; i++ {
		v1 := parsePHPOrFatal(t, testParsePHPOrderInputs[i])
		v2 := parsePHPOrFatal(t, testParsePHPOrderInputs[i+1])
		assert.True(
			t,
			Compare(v1, v2) < 0,
			"%v should be less than %v",
			testParsePHPOrderInputs[i],
			testParsePHPOrderInputs[i+1],
		)
	}
}

func parsePHPOrFatal(t *testing.T, v string) *Version {
	ver, err := ParsePHP(v)
	require.NoError(t, err, "no error parsing %v as a php version", v)
	return ver
}
