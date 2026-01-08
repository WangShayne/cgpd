package security

import (
	"encoding/json"
	"errors"
	"os"
	"time"
)

type Baseline struct {
	Version     int                      `json:"version"`
	GeneratedAt string                   `json:"generated_at"`
	Findings    map[string]BaselineEntry `json:"findings"`
}

type BaselineEntry struct {
	Kind    string `json:"kind"`
	RuleID  string `json:"rule_id"`
	File    string `json:"file"`
	Line    int    `json:"line"`
	Preview string `json:"preview"`
}

func LoadBaseline(path string) (Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Baseline{Version: 1, Findings: map[string]BaselineEntry{}}, nil
		}
		return Baseline{}, err
	}

	var b Baseline
	if err := json.Unmarshal(data, &b); err != nil {
		return Baseline{}, err
	}
	if b.Version == 0 {
		b.Version = 1
	}
	if b.Findings == nil {
		b.Findings = map[string]BaselineEntry{}
	}
	return b, nil
}

func WriteBaseline(path string, findings []Finding) error {
	b := Baseline{
		Version:     1,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Findings:    map[string]BaselineEntry{},
	}

	for _, f := range findings {
		if f.Fingerprint == "" {
			continue
		}
		b.Findings[f.Fingerprint] = BaselineEntry{
			Kind:    f.Kind,
			RuleID:  f.RuleID,
			File:    f.File,
			Line:    f.Line,
			Preview: f.Preview,
		}
	}

	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	return os.WriteFile(path, data, 0o644)
}

func NewFindingsSinceBaseline(findings []Finding, baseline Baseline) []Finding {
	if baseline.Findings == nil {
		return findings
	}

	out := make([]Finding, 0, len(findings))
	for _, f := range findings {
		if f.Fingerprint == "" {
			continue
		}
		if _, ok := baseline.Findings[f.Fingerprint]; ok {
			continue
		}
		out = append(out, f)
	}
	return out
}
