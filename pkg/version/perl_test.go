package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePerl(t *testing.T) {
	tests := map[ParsedAs]map[string]struct {
		version  string
		expected []string
	}{
		PerlDecimal: {
			"Text Is Invalid": {
				version: "1a",
			},
			"Decimal 1": {
				version: "1", expected: []string{"1", "0", "0"},
			},
			"Decimal 1.": {
				version: "1.", expected: []string{"1", "0", "0"},
			},
			"Decimal .2": {
				version: ".2", expected: []string{"0", "200", "0"},
			},
			"Decimal 1.2": {
				version: "1.2", expected: []string{"1", "200", "0"},
			},
			"Decimal 1.02": {
				version: "1.02", expected: []string{"1", "20", "0"},
			},
			"Decimal 1.002": {
				version: "1.002", expected: []string{"1", "2", "0"},
			},
			"Decimal 1.0023": {
				version: "1.0023", expected: []string{"1", "2", "300"},
			},
			"Decimal 1.00203": {
				version: "1.00203", expected: []string{"1", "2", "30"},
			},
			"Decimal 1.002003": {
				version: "1.002003", expected: []string{"1", "2", "3"},
			},
			"Decimal 1.00200304": {
				version: "1.00200304", expected: []string{"1", "2", "3", "40"},
			},
			"Decimal 1.00200": {
				version: "1.00200", expected: []string{"1", "2", "0"},
			},
			"Decimal Alpha Part Only Is Invalid": {
				version: "_123",
			},
			"Decimal Alpha Part Without Decimal Is Invalid": {
				version: "1_234",
			},
			"Decimal Alpha Part Without Fraction Digits Is Invalid": {
				version: "1._234",
			},
			"Decimal 1.0_2": {
				version: "1.0_2", expected: []string{"1", "20", "0"},
			},
			"Decimal 82.2_4568": {
				version: "82.2_4568", expected: []string{"82", "245", "680"},
			},
			"Decimal 01.02": {
				version: "01.02", expected: []string{"1", "20", "0"},
			},
		},
		PerlVString: {
			"v Only Is Invalid": {
				version: "v",
			},
			"Dotted Decimal v1": {
				version: "v1", expected: []string{"1", "0", "0"},
			},
			"Dotted Decimal v1.": {
				version: "v1.", expected: []string{"1", "0", "0"},
			},
			"Dotted Decimal .1.2": {
				version: ".1.2", expected: []string{"0", "1", "2"},
			},
			"v Without Integer Part Is Invalid": {
				version: "v.1.2",
			},
			"Dotted Decimal v1.2": {
				version: "v1.2", expected: []string{"1", "2", "0"},
			},
			"Dotted Decimal v1.2345": {
				version: "v1.2345", expected: []string{"1", "2345", "0"},
			},
			"Dotted Decimal v1.2.3": {
				version: "v1.2.3", expected: []string{"1", "2", "3"},
			},
			"Dotted Decimal v1.2.3.4": {
				version: "v1.2.3.4", expected: []string{"1", "2", "3", "4"},
			},
			"Dotted Decimal Alpha Part Only Is Invalid": {
				version: "v_123",
			},
			"Dotted Decimal Alpha Part Without Decimal Is Invalid": {
				version: "v1_234",
			},
			"Dotted Decimal Alpha Part Without Fraction Digits Is Invalid": {
				version: "v1._234",
			},
			"Dotted Decimal v1.0_2": {
				version: "v1.0_2", expected: []string{"1", "2", "0"},
			},
			"Dotted Decimal v1.02": {
				version: "v1.02", expected: []string{"1", "2", "0"},
			},
		},
	}

	for pa, cases := range tests {
		for name, tt := range cases {
			t.Run(name, func(t *testing.T) {
				actual, err := ParsePerl(tt.version)
				if tt.expected == nil {
					assert.Error(t, err)
				} else {
					require.NoError(t, err)
					assert.Equal(t, pa, actual.ParsedAs, "got expected ParsedAs value")
					assertDecimalEqualString(t, tt.expected, actual.Decimal)
					assertDecimalEqualDecimal(t, tt.expected, actual.Decimal)
				}
			})
		}
	}
}
