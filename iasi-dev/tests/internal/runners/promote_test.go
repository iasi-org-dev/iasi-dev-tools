package runners

import "testing"

func TestParseSemanticVersion(t *testing.T) {
	version, ok := parseSemanticVersion("1.2.3")
	if !ok || version.major != 1 || version.minor != 2 || version.patch != 3 {
		t.Fatalf("unexpected version: %+v, %t", version, ok)
	}
}

func TestParseSemanticVersionRejectsInvalidValues(t *testing.T) {
	for _, value := range []string{"v1.2.3", "v1.2", "01.2.3", "v1.x.3"} {
		if _, ok := parseSemanticVersion(value); ok {
			t.Fatalf("%q accepted", value)
		}
	}
}

func TestCompareSemanticVersions(t *testing.T) {
	a, _ := parseSemanticVersion("1.2.3")
	b, _ := parseSemanticVersion("1.3.0")
	if compareSemanticVersions(a, b) >= 0 {
		t.Fatal("expected a < b")
	}
}
