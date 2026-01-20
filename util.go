package main

import (
	"os"
	"regexp"
	"strings"
)

func normalizeAndCap(in []Finding, max int) []Finding {
	seen := map[string]bool{}
	out := make([]Finding, 0, len(in))
	for _, f := range in {
		key := f.Source + "|" + f.Message
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, f)
		if len(out) >= max {
			break
		}
	}
	return out
}

func safeName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "host"
	}
	re := regexp.MustCompile(`[^a-zA-Z0-9._-]+`)
	return re.ReplaceAllString(s, "_")
}

func nz(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "-"
	}
	return s
}

func shellEscape(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}
