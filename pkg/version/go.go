package version

import (
    "fmt"
    "regexp"
    "strings"
)

// 14 digits of YYYYMMDDhhmmss-12 hex digits for commit
var golangCommitSuffixRegEx = regexp.MustCompile(`([-\.]\d{14})-[a-f0-9]{12}$`)

func ParseGo(version string) (*Version, error) {
    version, err := normalizeGo(version)
    if err != nil {
        return nil, err
    }

    v, err := ParseGeneric(version)
    if err != nil {
        return nil, err
    }
    v.ParsedAs = Go
    return v, nil
}

func normalizeGo(version string) (string, error) {
    trimmed := strings.TrimPrefix(version, "v")

    if trimmed == version {
        return "", fmt.Errorf("Invalid Go version: %v", version)
    }

    trimmedSuffix := golangCommitSuffixRegEx.ReplaceAllString(trimmed, "${1}")

    return trimmedSuffix, nil
}
