package main

import (
	"encoding/csv"
	"encoding/json"
	"os"
)

func writeJSON(path string, rep Report) {
	b, err := json.MarshalIndent(rep, "", "  ")
	must(err)
	must(os.WriteFile(path, b, 0o644))
}

func writeCSV(path string, findings []Finding) {
	f, err := os.Create(path)
	must(err)
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	_ = w.Write([]string{"source", "severity", "when", "message"})
	for _, fi := range findings {
		_ = w.Write([]string{fi.Source, fi.Severity, fi.When, fi.Message})
	}
}
