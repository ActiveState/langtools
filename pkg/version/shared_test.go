package version

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/ericlagergren/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGeneric(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected []string
	}{
		{"Numbers", "0", []string{"0"}},
		{"Numbers", "1", []string{"1"}},
		{"Numbers", "1.0", []string{"1"}},
		{"Numbers", "0.92", []string{"0", "92"}},
		{"Numbers", "1-1.2", []string{"1", "1", "2"}},
		{"Sequential Dots", "1..2", []string{"1", "2"}},
		{"Sequential Dashes", "1--2", []string{"1", "2"}},
		{"Sequential Dot Dash", "1.-2", []string{"1", "2"}},
		{"Uppercase A", "A1", []string{"65", "1"}},
		{"Lowercase a", "a1", []string{"97", "1"}},
		{"Single Unicode", "小1", []string{"23567", "1"}},
		{"Ascii Word", "1.0bet", []string{"1", "0", "98.00000001010000000116"}},
		{"Unicode Word", "小寸-1.1", []string{"23567.0000023544", "1", "1"}},
		{"Unicode Separators", "1 2\u20013\u2002\u20034", []string{"1", "2", "3", "4"}},
		{"Normalizes Unicode", "e\u0301", []string{"233"}},
		{
			"Splits On Space",
			"10 Generic 142910-17",
			[]string{
				"10",
				"71.000000010100000001100000000101000000011400000001050000000099",
				"142910",
				"17",
			},
		},
		{"Drops Leading Zeros", "100.02.01", []string{"100", "2", "1"}},
		{"Pre-Release Identifier", "1.0-alpha", []string{"1", "0", "-26"}},
		{"Pre-Release Identifier Ignores Case", "1.0-AlPHa", []string{"1", "0", "-26"}},
		{"Pre-Release Identifier In Middle", "1.0-alpha.1", []string{"1", "0", "-26", "1"}},
		{"2 Pre-Release Identifiers", "1.0-alpha.beta", []string{"1", "0", "-26", "-25"}},
		{"Pre-Release Identifier Beta", "1.0-beta", []string{"1", "0", "-25"}},
		{"Pre-Release Identifier RC", "1.0-rc", []string{"1", "0", "-1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := ParseGeneric(tt.version)
			require.NoError(t, err)
			assert.Equal(t, Generic, actual.ParsedAs, "got expected ParsedAs value")
			assertStringEquality(t, tt.expected, actual)
			assertNumericEquality(t, tt.expected, actual)
		})
	}
}

func TestParseGenericPreReleaseIdentifierSortsCorrectly(t *testing.T) {
	alphaBeta := parseOrFatalGeneric(t, "1.0.0-alpha.beta")
	alpha := parseOrFatalGeneric(t, "1.0.0-alpha")
	alpha1 := parseOrFatalGeneric(t, "1.0.0-alpha.1")
	beta := parseOrFatalGeneric(t, "1.0.0-beta")
	beta2 := parseOrFatalGeneric(t, "1.0.0-beta.2")
	beta11 := parseOrFatalGeneric(t, "1.0.0-beta.11")
	rc := parseOrFatalGeneric(t, "1.0.0-rc.1")
	stable := parseOrFatalGeneric(t, "1.0.0")
	two0 := parseOrFatalGeneric(t, "2.0")
	two00 := parseOrFatalGeneric(t, "2.0.0")

	assert.True(t, Compare(alphaBeta, alpha) < 0, "Compare(alphaBeta, alpha)")
	assert.True(t, Compare(alpha, alpha1) < 0, "Compare(alpha, alpha1)")
	assert.True(t, Compare(alpha1, beta) < 0, "Compare(alpha1, beta)")
	assert.True(t, Compare(beta, beta2) < 0, "Compare(beta, beta2)")
	assert.True(t, Compare(beta2, beta11) < 0, "Compare(beta2, beta11)")
	assert.True(t, Compare(beta11, rc) < 0, "Compare(beta11, rc)")
	assert.True(t, Compare(rc, stable) < 0, "Compare(rc, stable)")
	assert.True(t, Compare(two0, two00) == 0, "Compare(two0, two00)")
}

func TestParseGenericParsesOpenSSLVersionsCorrectly(t *testing.T) {
	pre1 := parseOrFatalGeneric(t, "1.1.0-pre1")
	pre2 := parseOrFatalGeneric(t, "1.1.0-pre2")
	pre3 := parseOrFatalGeneric(t, "1.1.0-pre3")
	base := parseOrFatalGeneric(t, "1.1.0")
	baseA := parseOrFatalGeneric(t, "1.1.0a")
	baseB := parseOrFatalGeneric(t, "1.1.0b")
	baseC := parseOrFatalGeneric(t, "1.1.0c")

	assert.True(t, Compare(pre1, pre2) < 0)
	assert.True(t, Compare(pre2, pre3) < 0)
	assert.True(t, Compare(pre3, base) < 0)
	assert.True(t, Compare(base, baseA) < 0)
	assert.True(t, Compare(baseA, baseB) < 0)
	assert.True(t, Compare(baseB, baseC) < 0)
}

func TestParseSemVer(t *testing.T) {
	tests := map[string]struct {
		version  string
		expected []string
	}{
		"One Section Is Error": {
			version:  "1",
			expected: []string{},
		},
		"Two Sections Is Error": {
			version:  "1.0",
			expected: []string{},
		},
		"Number cannot have leading zero": {
			version:  "01.2.3",
			expected: []string{},
		},
		"Another invalid input": {
			version:  "0.0.0-.",
			expected: []string{},
		},
		"Parses Major.Minor.Patch": {
			version:  "1.2.3",
			expected: []string{"1", "2", "3"},
		},
		"Parses PreReleaseIdentifer": {
			version:  "1.2.3-a.1",
			expected: []string{"1", "2", "3", "-1", "97", "0", "1", "-1"},
		},
		"Parses alpha as pre-release": {
			version:  "1.2.3-alpha",
			expected: []string{"1", "2", "3", "-1", "97.108112104097", "-1"},
		},
		"Build Metadata Is Ignored": {
			version:  "1.2.3+ignored",
			expected: []string{"1", "2", "3"},
		},
		"Parses When All Sections Present": {
			version:  "1.2.3-a.1+ignored",
			expected: []string{"1", "2", "3", "-1", "97", "0", "1", "-1"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual, err := ParseSemVer(test.version)
			if len(test.expected) == 0 {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, SemVer, actual.ParsedAs, "got expected ParsedAs value")
				assertStringEquality(t, test.expected, actual)
				assertNumericEquality(t, test.expected, actual)
			}
		})
	}
}

var testParseSemVerOrderInputs = []string{
	"0.0.0-foo",
	"0.0.0",
	"0.0.1",
	"0.1.2",
	"0.9.0",
	"0.9.9",
	"0.10.0",
	"0.99.0",
	"1.0.0-alpha",
	"1.0.0-alpha.0",
	"1.0.0-alpha.1",
	"1.0.0-alpha.100",
	"1.0.0-alpha.100.0",
	"1.0.0-alpha.100.a",
	"1.0.0-alpha.beta",
	"1.0.0-beta",
	"1.0.0-beta.2",
	"1.0.0-beta.11",
	"1.0.0-rc.1",
	"1.0.0",
	"1.0.1",
	"1.2.2",
	"1.2.3-4",
	"1.2.3-5",
	"1.2.3-4-foo",
	"1.2.3-5-Foo",
	"1.2.3-5-foo",
	"1.2.3-R2",
	"1.2.3-a",
	"1.2.3-a.0",
	"1.2.3-a.5",
	"1.2.3-a.10",
	"1.2.3-a.100",
	"1.2.3-a.b",
	"1.2.3-a.b.c.5.d.100",
	"1.2.3-a.b.c.10.d.5",
	"1.2.3-alpha.0.2",
	"1.2.3-alpha.0.pr.1",
	"1.2.3-alpha.0.pr.2",
	"1.2.3-asdf",
	"1.2.3-pre",
	"1.2.3-r100",
	"1.2.3-r2",
	"1.2.3",
	"1.2.4-1",
	"1.2.4",
	"2.0.0",
	"2.3.4",
	"2.7.2+asdf",
	"3.0.0",
	"9.9.9-alpha.0.pr.1",
}

func TestParseSemVerOrdering(t *testing.T) {
	for i := 0; i < len(testParseSemVerOrderInputs)-1; i++ {
		v1 := parseOrFatalSemVer(t, testParseSemVerOrderInputs[i])
		v2 := parseOrFatalSemVer(t, testParseSemVerOrderInputs[i+1])
		assert.True(
			t,
			Compare(v1, v2) < 0,
			"%v should be less than %v",
			testParseSemVerOrderInputs[i],
			testParseSemVerOrderInputs[i+1],
		)
	}
}

func TestIsNumber(t *testing.T) {
	assert.True(t, isNumber("1"))
	assert.True(t, isNumber("1.0"))
	assert.True(t, isNumber("0.9"))
	assert.True(t, isNumber(".123"))

	assert.False(t, isNumber("a"))
	assert.False(t, isNumber("a1"))
	assert.False(t, isNumber("1a"))
	assert.False(t, isNumber("1.2.3"))
}

func assertStringEquality(t *testing.T, expected []string, actual *Version) {
	if actual.Decimal != nil {
		assertDecimalEqualString(t, expected, actual.Decimal)
	} else {
		assertIntsEqualString(t, expected, actual.Ints)
	}
}

func assertNumericEquality(t *testing.T, expected []string, actual *Version) {
	if actual.Decimal != nil {
		assertDecimalEqualDecimal(t, expected, actual.Decimal)
	} else {
		assertIntsEqualInts(t, expected, actual.Ints)
	}
}

func assertIntsEqualString(t *testing.T, expected []string, actual []int64) {
	require.Equal(t, len(expected), len(actual))
	for i := range expected {
		assert.Equal(t, expected[i], strconv.FormatInt(actual[i], 10))
	}
}

func assertDecimalEqualString(t *testing.T, expected []string, actual []*decimal.Big) {
	require.Equal(t, len(expected), len(actual))
	for i := range expected {
		assert.Equal(t, expected[i], actual[i].String())
	}
}

func assertIntsEqualInts(t *testing.T, expected []string, actual []int64) {
	expectedInts, err := stringsToInts(expected)
	assert.NoError(t, err)
	assert.Equal(t, expectedInts, actual)
}

func assertDecimalEqualDecimal(t *testing.T, expected []string, actual []*decimal.Big) {
	expectedDecimals, err := stringsToDecimals(expected)
	assert.NoError(t, err)
	assert.Equal(t, expectedDecimals, actual)
}

type Cmp uint8

const (
	LT Cmp = iota
	EQ
	GT
)

func TestCompare(t *testing.T) {
	type testCase struct {
		v1, v2 *Version
		expect Cmp
	}

	testCases := map[string]testCase{
		"less than one segment": {
			v1:     parseOrFatalGeneric(t, "1"),
			v2:     parseOrFatalGeneric(t, "2"),
			expect: LT,
		},
		"less than two segments": {
			v1:     parseOrFatalGeneric(t, "3.abc"),
			v2:     parseOrFatalGeneric(t, "3.def"),
			expect: LT,
		},
		"less than three segments": {
			v1:     parseOrFatalGeneric(t, "0.1.78"),
			v2:     parseOrFatalGeneric(t, "0.2.78"),
			expect: LT,
		},
		"less than different length": {
			v1:     parseOrFatalGeneric(t, "1.0"),
			v2:     parseOrFatalGeneric(t, "1.0.1"),
			expect: LT,
		},
		"equal one segment": {
			v1:     parseOrFatalGeneric(t, "1"),
			v2:     parseOrFatalGeneric(t, "1"),
			expect: EQ,
		},
		"equal two segments": {
			v1:     parseOrFatalGeneric(t, "3.abc"),
			v2:     parseOrFatalGeneric(t, "3.abc"),
			expect: EQ,
		},
		"equal three segments": {
			v1:     parseOrFatalGeneric(t, "0.2.78"),
			v2:     parseOrFatalGeneric(t, "0.2.78"),
			expect: EQ,
		},
		"greater than one segment": {
			v1:     parseOrFatalGeneric(t, "10"),
			v2:     parseOrFatalGeneric(t, "1"),
			expect: GT,
		},
		"greater than two segments": {
			v1:     parseOrFatalGeneric(t, "1.101"),
			v2:     parseOrFatalGeneric(t, "1.10"),
			expect: GT,
		},
		"greater than three segments": {
			v1:     parseOrFatalGeneric(t, "4.8.23abd"),
			v2:     parseOrFatalGeneric(t, "4.8.23abc"),
			expect: GT,
		},
		"greater than different length": {
			v1:     parseOrFatalGeneric(t, "0"),
			v2:     parseOrFatalGeneric(t, "0.0.23"),
			expect: LT,
		},
		"equal different length (first is longer)": {
			v1:     parseOrFatalGeneric(t, "1.1.2.0"),
			v2:     parseOrFatalGeneric(t, "1.1.2"),
			expect: EQ,
		},
		"equal different length (second is longer)": {
			v1:     parseOrFatalGeneric(t, "1.1.2"),
			v2:     parseOrFatalGeneric(t, "1.1.2.0"),
			expect: EQ,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			switch testCase.expect {
			case LT:
				assert.Truef(t, Compare(testCase.v1, testCase.v2) < 0, "%s is less than %s", testCase.v1, testCase.v2)
			case EQ:
				assert.Equalf(t, 0, Compare(testCase.v1, testCase.v2), "%s is equal to %s", testCase.v1, testCase.v2)
			case GT:
				assert.Truef(t, Compare(testCase.v1, testCase.v2) > 0, "%s is greater than %s", testCase.v1, testCase.v2)
			}
		})
	}
}

func TestClone(t *testing.T) {
	v1 := parseOrFatalGeneric(t, "1.2")
	v2 := v1.Clone()

	assert.Equal(t, 0, Compare(v1, v2), "cloned version has same Decimal slice")
	assert.Equal(t, v1.Original, v2.Original, "cloned version has same Original string")
	assert.Equal(t, v1.ParsedAs, v2.ParsedAs, "cloned version has same ParsedAs value")

	v1.Ints[0] = 0
	assert.NotEqual(t, 0, Compare(v1, v2), "changing slice in original does not change clone")
}

func TestString(t *testing.T) {
	v := parseOrFatalGeneric(t, "1.2")
	assert.Equal(t, "1.2 (Generic)", v.String())

	v = parseOrFatalSemVer(t, "1.2.3")
	assert.Equal(t, "1.2.3 (SemVer)", v.String())
}

func TestTrimTrailingZeros(t *testing.T) {
	tests := []struct {
		input, expected []string
	}{
		{[]string{"0"}, []string{"0"}},
		{[]string{"1"}, []string{"1"}},
		{[]string{"0", "0"}, []string{"0"}},
		{[]string{"0", "1"}, []string{"0", "1"}},
		{[]string{"1", "0"}, []string{"1"}},
		{[]string{"1", "1"}, []string{"1", "1"}},
		{[]string{"0", "0", "0"}, []string{"0"}},
		{[]string{"1", "0", "0"}, []string{"1"}},
		{[]string{"1", "0", "1"}, []string{"1", "0", "1"}},
		{[]string{"1", "1", "1"}, []string{"1", "1", "1"}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.input), func(t *testing.T) {
			input := mustStringsToDecimal(t, tt.input)
			actual := trimTrailingZerosDecimals(input)
			expected := mustStringsToDecimal(t, tt.expected)
			assert.Equal(t, expected, actual, "expected %v got %v", expected, actual)
		})
	}
}

func parseOrFatalGeneric(t *testing.T, v string) *Version {
	ver, err := ParseGeneric(v)
	assert.NoError(t, err, "no error parsing %s as a generic version", v)

	return ver
}

func parseOrFatalSemVer(t *testing.T, v string) *Version {
	ver, err := ParseSemVer(v)
	assert.NoError(t, err, "no error parsing %s as a semver version", v)

	return ver
}

func mustStringsToDecimal(t *testing.T, s []string) []*decimal.Big {
	d, err := stringsToDecimals(s)
	assert.NoError(t, err, "no error parsing strings to decimals")

	return d
}
