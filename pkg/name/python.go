// Package name provides name normalization for package names
package name

import (
	"regexp"
	"strings"
)

var replacement = regexp.MustCompile(`[\.\_-]+`)

// NormalizePython takes a Python package name and returns it in normalized
// form. Specifically, that means it is in all lower case and all periods (.)
// and underscores (_) with hyphens. See
// https://www.python.org/dev/peps/pep-0503/#normalized-names for details on
// how names should be normalized in Python.
func NormalizePython(name string) string {
	return strings.ToLower(replacement.ReplaceAllString(name, "-"))
}
