// Package version provides parsing of version strings, such as "1.2.3". The
// package returns a struct that contains several pieces of information.
//
// The primary motivation for this package was to create a representation of
// versions that can be stored and sorted in a Postgres database. To that end,
// we turn all versions into a slice of `*decimal.Big`
// (github.com/ericlagergren/decimal) values.
//
// It is not possible to produce reasonably sortable versions across multiple
// language ecosystems, or even between different packages in the same
// ecosystem. Instead, we simply aim to make sure that all versions of a
// single package are sortable, even if the versioning scheme for that package
// changes over time. We assume that even if the versioning scheme changes,
// that the "major" version portion of the new scheme will sort higher than
// the "major" version of the old scheme. We have "major" in quotes there
// because some versioning schemes don't really distinguish between major,
// minor, and patch portions of a version.
//
// The various parsing schemes handle non-numeric components as well as
// numbers appropriately. For some schemes, non-numeric components like
// strings have special meaning. For example, in semver there is a strict
// ordering for "alpha", "beta", etc. For other schemes we simply encode
// non-numeric values into decimal values using the Unicode codepoint for each
// letter.
//
// We currently make no guarantee that the decimal representation of a version
// will not change between releases of this module. That means that if you
// store the versions you may have to re-parse existing versions if you
// upgrade this module. Given that fact, you should always make sure to store
// the original version that you parsed alongside the decimal slice
// representation.
package version

//go:generate enumer -type ParsedAs .

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ericlagergren/decimal"
)

// ParsedAs is an enum for the various types of versions we can parse.
type ParsedAs int

const (
	// Unknown should never be used.
	Unknown ParsedAs = iota
	// Generic is a generic version.
	Generic
	// SemVer is the well known semver scheme (https://semver.org/).
	SemVer
	// PerlDecimal is for Perl versions which are simply numbers (42, 1.2, etc.).
	PerlDecimal
	// PerlVString is for Perl v-strings like "v1.1.2" (but these don't
	// require the leading "v", so "1.2.3" is also valid).
	PerlVString
	// PHP is for PHP versions as used by composer.
	PHP
	// PythonLegacy is for Python versions before PEP440 was adopted.
	PythonLegacy
	// PythonPEP440 is for versions as described in PEP440.
	PythonPEP440
	// Ruby is for Ruby versions.
	Ruby
)

// Version is the struct returned from all parsing funcs.
type Version struct {
	// Original is the string that was passed to the parsing func.
	Original string `json:"version"`
	// The simplest form of data
	Ints []int64
	// Decimal contains a slice of `*decimal.Big` values. This will always
	// contain at least one element.
	Decimal []*decimal.Big `json:"sortable_version"`
	// ParsedAs indicates which type the version was parsed as.
	ParsedAs ParsedAs `json:"-"`
}

func (v *Version) MarshalJSON() ([]byte, error) {
	if v.Decimal != nil {
		return json.Marshal(&struct {
			Original string         `json:"version"`
			Decimal  []*decimal.Big `json:"sortable_version"`
			ParsedAs ParsedAs       `json:"-"`
		}{
			Original: v.Original,
			Decimal:  v.Decimal,
			ParsedAs: v.ParsedAs,
		})
	}
	return json.Marshal(&struct {
		Original string   `json:"version"`
		Ints     []int64  `json:"sortable_version"`
		ParsedAs ParsedAs `json:"-"`
	}{
		Original: v.Original,
		Ints:     v.Ints,
		ParsedAs: v.ParsedAs,
	})
}

// fromStringSlice take a version type and a slice of strings and returns a
// new Version struct. Each element of the string slice should contain a
// string representation of a number. This returns an error if any element of
// the slice cannot be converted to a *decimal.Big value.
func fromStringSlice(pa ParsedAs, original string, strings []string) (*Version, error) {
	decimals, err := stringsToDecimals(strings)
	if err != nil {
		return nil, err
	}

	decimals = trimTrailingZeros(decimals)
	return &Version{
		Original: original,
		Decimal:  decimals,
		ParsedAs: pa,
	}, nil
}

func stringsToDecimals(strings []string) ([]*decimal.Big, error) {
	if len(strings) == 0 {
		return nil, errors.New("The provided string slice must have at least one element")
	}

	decimals := make([]*decimal.Big, len(strings))
	for i, s := range strings {
		d := &decimal.Big{}
		if _, ok := d.SetString(s); !ok {
			return nil, errors.New("Failed to create decimal.Big from " + s)
		}
		decimals[i] = d
	}

	return decimals, nil
}

func trimTrailingZeros(decimals []*decimal.Big) []*decimal.Big {
	indexOfLastZero := len(decimals)
	for i := len(decimals) - 1; i > 0; i-- {
		if decimals[i].Cmp(bigZero) != 0 {
			break
		}
		indexOfLastZero = i
	}

	return decimals[0:indexOfLastZero]
}

var bigZero = decimal.New(0, 0)

// Compare returns:
//   <0 if the version in v1 is less than the version in v2
//    0 if the version in v1 is equal to the version in v2
//   >0 if the version in v1 is greater than the version in v2
//
// Versions that differ only by trailing zeros (e.g. "1.2" and "1.2.0") are
// equal.
func Compare(v1, v2 *Version) int {
	min, max, longest, flip := minMax(v1.Decimal, v2.Decimal)

	// find any difference between these versions where they have the same number of segments
	for i := 0; i < min; i++ {
		cmp := v1.Decimal[i].Cmp(v2.Decimal[i])
		if cmp != 0 {
			return cmp
		}
	}

	// compare remaining segments to zero
	for i := min; i < max; i++ {
		cmp := longest[i].Cmp(bigZero)
		if cmp != 0 {
			return cmp * flip
		}
	}

	return 0
}

// helper function to find the lengths of and longest version segment array
func minMax(v1 []*decimal.Big, v2 []*decimal.Big) (int, int, []*decimal.Big, int) {
	l1 := len(v1)
	l2 := len(v2)

	if l1 < l2 {
		return l1, l2, v2, -1
	}
	return l2, l1, v1, 1
}

// Clone returns a new *Version that is a clone of the one passed as the
// method receiver.
func (v *Version) Clone() *Version {
	d := make([]*decimal.Big, len(v.Decimal))
	for i := range v.Decimal {
		d[i] = decimal.New(0, 0)
		d[i].Copy(v.Decimal[i])
	}
	return &Version{
		Original: v.Original,
		Decimal:  d,
		ParsedAs: v.ParsedAs,
	}
}

// String returns a string representation of the version. Note that this is
// not the same as v.Original.
func (v *Version) String() string {
	return fmt.Sprintf("%s (%s)", v.Original, v.ParsedAs.String())
}
