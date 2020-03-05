package name

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizePython(t *testing.T) {
	cases := map[string]string{
		"flask":                                  "flask",
		"Flask":                                  "flask",
		"FLASK":                                  "flask",
		"backports.ssl":                          "backports-ssl",
		"backports-----ssl":                      "backports-ssl",
		"backports.SSL":                          "backports-ssl",
		"Backports.SSL":                          "backports-ssl",
		"backports-datetime-fromisoformat":       "backports-datetime-fromisoformat",
		"backports-datetime_fromisoformat":       "backports-datetime-fromisoformat",
		"BACKPORTS-DATETIME-FROMISOFORMAT":       "backports-datetime-fromisoformat",
		"BACKPORTS-.-DATETIME__-.-FROMISOFORMAT": "backports-datetime-fromisoformat",
	}

	for from, norm := range cases {
		assert.Equal(t, norm, NormalizePython(from), `normalization of "%s" is "%s"`, from, norm)
	}
}
