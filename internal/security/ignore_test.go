package security

import (
	"os"
	"testing"
)

func TestLoadIgnoreFileAndMatch(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "cgpd-ignore-*")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	if _, err := f.WriteString("# comment\npath:secrets/*\nrule:github-token\nfingerprint:abcd\n*.pem\n"); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	rules, err := LoadIgnoreFile(f.Name())
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if !rules.MatchesPath("secrets/a.txt") {
		t.Fatalf("expected path match")
	}
	if !rules.MatchesPath("x.pem") {
		t.Fatalf("expected basename glob match")
	}

	if !rules.MatchesFinding(Finding{RuleID: "github-token"}) {
		t.Fatalf("expected rule match")
	}
	if !rules.MatchesFinding(Finding{Fingerprint: "ABCD"}) {
		t.Fatalf("expected fingerprint match")
	}
}
