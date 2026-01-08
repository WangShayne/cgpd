package security

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type Finding struct {
	Kind        string `json:"kind"`
	File        string `json:"file"`
	RuleID      string `json:"rule_id"`
	Fingerprint string `json:"fingerprint"`
	Line        int    `json:"line"`
	Preview     string `json:"preview"`
}

type diffRule struct {
	id           string
	kind         string
	re           *regexp.Regexp
	captureGroup int
	minLen       int
	minEntropy   float64
}

var diffRules = []diffRule{
	{
		id:           "private-key-header",
		kind:         "secret",
		re:           regexp.MustCompile(`-----BEGIN [A-Z0-9 ]{0,32}PRIVATE KEY-----`),
		captureGroup: 0,
	},
	{
		id:           "openai-api-key",
		kind:         "secret",
		re:           regexp.MustCompile(`\bsk-[A-Za-z0-9]{20,}\b`),
		captureGroup: 0,
		minLen:       24,
	},
	{
		id:           "github-token",
		kind:         "secret",
		re:           regexp.MustCompile(`\bgh[pousr]_[A-Za-z0-9]{36,255}\b`),
		captureGroup: 0,
		minLen:       40,
	},
	{
		id:           "google-api-key",
		kind:         "secret",
		re:           regexp.MustCompile(`\bAIza[0-9A-Za-z\-_]{35}\b`),
		captureGroup: 0,
		minLen:       35,
	},
	{
		id:           "aws-access-key-id",
		kind:         "secret",
		re:           regexp.MustCompile(`\b(A3T[A-Z0-9]{16}|AKIA[0-9A-Z]{16}|ASIA[0-9A-Z]{16})\b`),
		captureGroup: 1,
	},
	{
		id:           "slack-token",
		kind:         "secret",
		re:           regexp.MustCompile(`\bxox[baprs]-[A-Za-z0-9-]{10,255}\b`),
		captureGroup: 0,
		minLen:       20,
	},
	{
		id:           "generic-assignment-high-entropy",
		kind:         "secret",
		re:           regexp.MustCompile(`(?i)\b(api[_-]?key|secret|token|password)\b\s*[:=]\s*['"]?([A-Za-z0-9+/=_\-]{12,})['"]?`),
		captureGroup: 2,
		minLen:       20,
		minEntropy:   3.6,
	},
}

func ScanStagedDiff(diff string) []Finding {
	var findings []Finding

	scanner := bufio.NewScanner(strings.NewReader(diff))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	currentFile := ""
	newLine := 0

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "diff --git ") {
			currentFile = parseDiffFile(line)
			newLine = 0
			continue
		}

		if strings.HasPrefix(line, "@@ ") {
			if start := parseHunkNewStart(line); start > 0 {
				newLine = start - 1
			}
			continue
		}

		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			continue
		}

		if strings.HasPrefix(line, "+") {
			if strings.HasPrefix(line, "+++") {
				continue
			}
			newLine++
			content := strings.TrimRight(line[1:], "\r")
			findings = append(findings, scanLine(currentFile, newLine, content)...)
			continue
		}

		if strings.HasPrefix(line, " ") {
			newLine++
		}
	}

	return findings
}

func scanLine(file string, line int, content string) []Finding {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return nil
	}

	var findings []Finding
	for _, rule := range diffRules {
		matches := rule.re.FindAllStringSubmatchIndex(content, -1)
		if len(matches) == 0 {
			continue
		}
		for _, m := range matches {
			value := extractCapture(content, m, rule.captureGroup)
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}
			if rule.minLen > 0 && len(value) < rule.minLen {
				continue
			}
			if rule.minEntropy > 0 {
				if e := shannonEntropy(value); e < rule.minEntropy {
					continue
				}
			}

			fp := fingerprint(rule.id, value)
			findings = append(findings, Finding{
				Kind:        rule.kind,
				File:        file,
				RuleID:      rule.id,
				Fingerprint: fp,
				Line:        line,
				Preview:     maskValue(value),
			})
		}
	}

	return uniqueFindings(findings)
}

func uniqueFindings(in []Finding) []Finding {
	seen := make(map[string]struct{}, len(in))
	out := make([]Finding, 0, len(in))
	for _, f := range in {
		key := f.RuleID + ":" + f.Fingerprint + ":" + f.File + ":" + strconv.Itoa(f.Line)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, f)
	}
	return out
}

func extractCapture(content string, idx []int, group int) string {
	if group == 0 {
		if len(idx) >= 2 && idx[0] >= 0 && idx[1] >= 0 {
			return content[idx[0]:idx[1]]
		}
		return ""
	}

	off := group * 2
	if len(idx) <= off+1 {
		return ""
	}
	start, end := idx[off], idx[off+1]
	if start < 0 || end < 0 || end < start {
		return ""
	}
	return content[start:end]
}

func fingerprint(ruleID string, value string) string {
	sum := sha256.Sum256([]byte(ruleID + ":" + value))
	return hex.EncodeToString(sum[:])
}

func maskValue(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "-----BEGIN ") {
		return s
	}
	if len(s) <= 8 {
		return "***"
	}
	return s[:3] + "..." + s[len(s)-2:]
}

func shannonEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}

	counts := make(map[rune]int)
	for _, r := range s {
		counts[r]++
	}

	var entropy float64
	invLen := 1.0 / float64(len([]rune(s)))
	for _, c := range counts {
		p := float64(c) * invLen
		entropy -= p * math.Log2(p)
	}
	return entropy
}

func parseDiffFile(line string) string {
	parts := strings.Fields(line)
	if len(parts) < 4 {
		return ""
	}
	b := parts[3]
	if strings.HasPrefix(b, "b/") {
		return strings.TrimPrefix(b, "b/")
	}
	return b
}

func parseHunkNewStart(line string) int {
	plus := strings.Index(line, "+")
	if plus == -1 {
		return 0
	}
	end := strings.Index(line[plus:], " ")
	if end == -1 {
		return 0
	}
	chunk := line[plus+1 : plus+end]
	chunk = strings.TrimPrefix(chunk, "+")
	comma := strings.Index(chunk, ",")
	if comma != -1 {
		chunk = chunk[:comma]
	}
	n, err := strconv.Atoi(chunk)
	if err != nil {
		return 0
	}
	return n
}
