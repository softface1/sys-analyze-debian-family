package main

import (
	"bytes"
	"html/template"
	"os"
)

func writeHTML(path string, rep Report) {
	tpl := template.Must(template.New("r").Parse(htmlTemplate))
	var buf bytes.Buffer
	must(tpl.Execute(&buf, rep))
	must(os.WriteFile(path, buf.Bytes(), 0o644))
}
