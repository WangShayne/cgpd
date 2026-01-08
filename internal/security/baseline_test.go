package security

import (
	"os"
	"testing"
)

func TestBaselineRoundTrip(t *testing.T) {
	path := t.TempDir() + "/baseline.json"
	findings := []Finding{
		{Kind: "secret", RuleID: "r1", File: "a.go", Line: 1, Preview: "abc", Fingerprint: "fp1"},
		{Kind: "secret", RuleID: "r2", File: "b.go", Line: 2, Preview: "def", Fingerprint: "fp2"},
	}

	if err := WriteBaseline(path, findings); err != nil {
		t.Fatalf("write: %v", err)
	}

	b, err := LoadBaseline(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if b.Version != 1 {
		t.Fatalf("unexpected version: %d", b.Version)
	}
	if len(b.Findings) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(b.Findings))
	}
	if _, ok := b.Findings["fp1"]; !ok {
		t.Fatalf("missing fp1")
	}

	newOnes := NewFindingsSinceBaseline([]Finding{{Fingerprint: "fp2"}, {Fingerprint: "fp3"}}, b)
	if len(newOnes) != 1 || newOnes[0].Fingerprint != "fp3" {
		t.Fatalf("unexpected new findings: %+v", newOnes)
	}

	if err := os.Remove(path); err != nil {
		t.Fatalf("remove: %v", err)
	}
}
