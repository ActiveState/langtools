package version

import (
	"testing"
)

func BenchmarkCompare(b *testing.B) {
	versions := []*Version{}
	for _, s := range pythonTestStrings {
		v, err := ParsePython(s)
		if err != nil {
			b.Fatal(err)
		}
		versions = append(versions, v)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, v1 := range versions {
			for _, v2 := range versions {
				Compare(v1, v2)
			}
		}
	}
}
