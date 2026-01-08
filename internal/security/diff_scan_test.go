package security

import "testing"

func TestScanStagedDiff(t *testing.T) {
	diff := "" +
		"diff --git a/main.go b/main.go\n" +
		"index 1111111..2222222 100644\n" +
		"--- a/main.go\n" +
		"+++ b/main.go\n" +
		"@@ -1,2 +1,4 @@\n" +
		" package main\n" +
		"+var openai = \"sk-ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcd\"\n" +
		"+var gh = \"ghp_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcd\"\n" +
		"+var hdr = \"-----BEGIN OPENSSH PRIVATE KEY-----\"\n"

	findings := ScanStagedDiff(diff)
	if len(findings) == 0 {
		t.Fatalf("expected findings, got 0")
	}

	seen := map[string]bool{}
	for _, f := range findings {
		if f.File != "main.go" {
			t.Fatalf("unexpected file %q", f.File)
		}
		if f.RuleID == "" || f.Fingerprint == "" {
			t.Fatalf("missing rule or fingerprint: %+v", f)
		}
		seen[f.RuleID] = true
	}

	for _, want := range []string{"openai-api-key", "github-token", "private-key-header"} {
		if !seen[want] {
			t.Fatalf("expected rule %q to match; got %v", want, seen)
		}
	}
}

func TestScanStagedDiff_IgnoresLowEntropyGenericAssignment(t *testing.T) {
	diff := "" +
		"diff --git a/app.txt b/app.txt\n" +
		"--- a/app.txt\n" +
		"+++ b/app.txt\n" +
		"@@ -1 +1 @@\n" +
		"+password = \"aaaaaaaaaaaaaaaaaaaa\"\n"

	findings := ScanStagedDiff(diff)
	for _, f := range findings {
		if f.RuleID == "generic-assignment-high-entropy" {
			t.Fatalf("unexpected generic high-entropy match: %+v", f)
		}
	}
}

func BenchmarkScanStagedDiff(b *testing.B) {
	line := "+var openai = \"sk-ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcd\"\n"
	diff := "diff --git a/main.go b/main.go\n" +
		"--- a/main.go\n" +
		"+++ b/main.go\n" +
		"@@ -1 +1 @@\n" +
		" package main\n"
	for i := 0; i < 2000; i++ {
		diff += line
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ScanStagedDiff(diff)
	}
}
