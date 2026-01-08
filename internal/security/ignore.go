package security

import (
	"bufio"
	"errors"
	"os"
	"path"
	"strings"
)

type IgnoreRules struct {
	PathGlobs    []string
	RuleIDs      map[string]struct{}
	Fingerprints map[string]struct{}
}

func LoadIgnoreFile(pathOrEmpty string) (IgnoreRules, error) {
	p := strings.TrimSpace(pathOrEmpty)
	if p == "" {
		return IgnoreRules{}, nil
	}

	f, err := os.Open(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return IgnoreRules{}, nil
		}
		return IgnoreRules{}, err
	}
	defer f.Close()

	rules := IgnoreRules{
		RuleIDs:      map[string]struct{}{},
		Fingerprints: map[string]struct{}{},
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}

		key, val, ok := strings.Cut(line, ":")
		if !ok {
			rules.PathGlobs = append(rules.PathGlobs, line)
			continue
		}

		key = strings.TrimSpace(strings.ToLower(key))
		val = strings.TrimSpace(val)
		if val == "" {
			continue
		}

		switch key {
		case "path":
			rules.PathGlobs = append(rules.PathGlobs, val)
		case "rule":
			rules.RuleIDs[val] = struct{}{}
		case "fingerprint":
			rules.Fingerprints[strings.ToLower(val)] = struct{}{}
		default:
			rules.PathGlobs = append(rules.PathGlobs, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return IgnoreRules{}, err
	}
	return rules, nil
}

func (r IgnoreRules) MatchesPath(p string) bool {
	p = strings.TrimSpace(p)
	if p == "" {
		return false
	}

	norm := path.Clean(strings.ReplaceAll(p, "\\", "/"))
	for _, g := range r.PathGlobs {
		g = strings.TrimSpace(g)
		if g == "" {
			continue
		}
		if ok, _ := path.Match(g, norm); ok {
			return true
		}
		if !strings.Contains(g, "/") {
			if ok, _ := path.Match(g, path.Base(norm)); ok {
				return true
			}
		}
	}
	return false
}

func (r IgnoreRules) MatchesFinding(f Finding) bool {
	if f.Fingerprint != "" {
		if _, ok := r.Fingerprints[strings.ToLower(f.Fingerprint)]; ok {
			return true
		}
	}
	if f.RuleID != "" {
		if _, ok := r.RuleIDs[f.RuleID]; ok {
			return true
		}
	}
	if f.File != "" {
		if r.MatchesPath(f.File) {
			return true
		}
	}
	return false
}

func FilterPaths(paths []string, rules IgnoreRules) []string {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		if rules.MatchesPath(p) {
			continue
		}
		out = append(out, p)
	}
	return out
}

func FilterFindings(findings []Finding, rules IgnoreRules) []Finding {
	out := make([]Finding, 0, len(findings))
	for _, f := range findings {
		if rules.MatchesFinding(f) {
			continue
		}
		out = append(out, f)
	}
	return out
}
