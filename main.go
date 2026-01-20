package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func main() {
	var cfg Config
	var timeoutSec int

	flag.StringVar(&cfg.OutDir, "outdir", "./snapshot_out", "output directory")
	flag.StringVar(&cfg.Since, "since", "24h", "time window for logs (e.g. 2h, 24h, 7d)")
	flag.BoolVar(&cfg.KeepRaw, "raw", true, "save raw command outputs into outdir/raw")
	flag.IntVar(&cfg.MaxFindings, "max", 2500, "max findings to keep (global cap)")
	flag.IntVar(&timeoutSec, "timeout", 12, "command timeout seconds")
	flag.BoolVar(&cfg.MakeHTML, "html", true, "generate report.html")
	flag.BoolVar(&cfg.DoPack, "pack", false, "pack outdir into tar.gz after generation")
	flag.Parse()

	cfg.CmdTimeout = time.Duration(timeoutSec) * time.Second

	must(os.MkdirAll(cfg.OutDir, 0o755))
	rawDir := filepath.Join(cfg.OutDir, "raw")
	if cfg.KeepRaw {
		must(os.MkdirAll(rawDir, 0o755))
	}

	rep := Report{Extra: map[string]any{}}
	rep.Meta.GeneratedAt = time.Now().Format(time.RFC3339)
	rep.Meta.Arch = runtime.GOARCH
	rep.Meta.OS = runtime.GOOS
	rep.Meta.Since = cfg.Since
	rep.Meta.User = os.Getenv("SUDO_USER")
	if rep.Meta.User == "" {
		rep.Meta.User = os.Getenv("USER")
	}

	ex := &Exec{
		Timeout:  cfg.CmdTimeout,
		KeepRaw:  cfg.KeepRaw,
		RawDir:   rawDir,
		Commands: &rep.Commands,
	}

	rep.Meta.Hostname = strings.TrimSpace(ex.Run("hostname").Stdout)
	rep.Meta.Kernel = strings.TrimSpace(ex.Run("uname -r").Stdout)

	for _, m := range defaultModules() {
		if err := m.Run(ex, &rep, cfg); err != nil {
			rep.Findings = append(rep.Findings, Finding{
				Source:   m.ID(),
				Severity: "WARN",
				Message:  "module error: " + err.Error(),
			})
		}
	}

	rep.Findings = normalizeAndCap(rep.Findings, cfg.MaxFindings)

	writeJSON(filepath.Join(cfg.OutDir, "report.json"), rep)
	writeCSV(filepath.Join(cfg.OutDir, "errors.csv"), rep.Findings)
	writeMD(filepath.Join(cfg.OutDir, "report.md"), rep)
	if cfg.MakeHTML {
		writeHTML(filepath.Join(cfg.OutDir, "report.html"), rep)
	}

	if cfg.DoPack {
		packPath := filepath.Join(cfg.OutDir, fmt.Sprintf("snapshot_%s_%s.tar.gz",
			safeName(rep.Meta.Hostname),
			time.Now().Format("20060102_150405"),
		))
		must(packDirTarGz(cfg.OutDir, packPath))
		fmt.Printf("OK: packed to %s\n", packPath)
	}

	fmt.Printf("OK: wrote report to %s\n", cfg.OutDir)
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "FATAL:", err)
		os.Exit(1)
	}
}
