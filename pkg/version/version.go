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
	// PythonLegacy is for Python versions before PEP440 was adopted.
	PythonLegacy
	// PythonPEP440 is for versions as described in PEP440.
	PythonPEP440
)

// Version is the struct returned from all parsing funcs.
type Version struct {
	// Original is the string that was passed to the parsing func.
	Original string `json:"version"`
	// Decimal contains a slice of `*decimal.Big` values. This will always
	// contain at least one element.
	Decimal []*decimal.Big `json:"sortable_version"`
	// ParsedAs indicates which type the version was parsed as.
	ParsedAs ParsedAs `json:"-"`
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

var bigZero = decimal.New(0, 0)

// Compare returns:
//   <0 if the version in v1 is less than the version in v2
//    0 if the version in v1 is equal to the version in v2
//   >0 if the version in v1 is greater than the version in v2
//
// If the two versions have different numbers of segments, and both end in
// zero ("1.0" and "1.0.0"), then they compare as equal. However, "1.1" >
// "1.0.0" and "0.9" < "1.0.0".
func Compare(v1, v2 *Version) int {
	if len(v1.Decimal) > len(v2.Decimal) {
		v2 = v2.Clone()
		for i := len(v2.Decimal); i < len(v1.Decimal); i++ {
			if v1.Decimal[i].Cmp(bigZero) != 0 {
				break
			}
			v2.Decimal = append(v2.Decimal, decimal.New(0, 0))
		}
	} else if len(v1.Decimal) < len(v2.Decimal) {
		v1 = v1.Clone()
		for i := len(v1.Decimal); i < len(v2.Decimal); i++ {
			if v2.Decimal[i].Cmp(bigZero) != 0 {
				break
			}
			v1.Decimal = append(v1.Decimal, decimal.New(0, 0))
		}
	}

	min := len(v1.Decimal)
	if len(v2.Decimal) < min {
		min = len(v2.Decimal)
	}

	for i := 0; i < min; i++ {
		if v1.Decimal[i].Cmp(v2.Decimal[i]) != 0 {
			return v1.Decimal[i].Cmp(v2.Decimal[i])
		}
	}

	return len(v1.Decimal) - len(v2.Decimal)
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
