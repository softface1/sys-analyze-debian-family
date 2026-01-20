package main

import (
	"fmt"
	"os"
	"strings"
)

func writeMD(path string, rep Report) {
	var b strings.Builder

	b.WriteString("# System Snapshot Report\n\n")
	b.WriteString(fmt.Sprintf("- Generated: `%s`\n", rep.Meta.GeneratedAt))
	b.WriteString(fmt.Sprintf("- Hostname: `%s`\n", rep.Meta.Hostname))
	b.WriteString(fmt.Sprintf("- Kernel: `%s`\n", rep.Meta.Kernel))
	b.WriteString(fmt.Sprintf("- Since: `%s`\n\n", rep.Meta.Since))

	b.WriteString("## Health\n\n")
	b.WriteString(fmt.Sprintf("- Score: **%d** / 100 (grade **%s**)\n", rep.Health.Score, rep.Health.Grade))
	for _, r := range rep.Health.Reasons {
		b.WriteString(fmt.Sprintf("  - %s\n", r))
	}
	b.WriteString("\n")

	b.WriteString("## Quick Summary\n\n")
	b.WriteString("### Uptime\n\n```\n" + safe(rep.Summary.Uptime) + "\n```\n\n")
	b.WriteString("### Loadavg\n\n```\n" + safe(rep.Summary.LoadAvg) + "\n```\n\n")
	b.WriteString("### Memory\n\n```\n" + safe(rep.Summary.MemInfo) + "\n```\n\n")
	b.WriteString("### Disk (df -hT)\n\n```\n" + safe(rep.Summary.DiskDf) + "\n```\n\n")
	b.WriteString("### Block devices (lsblk)\n\n```\n" + safe(rep.Summary.LsblkText) + "\n```\n\n")
	b.WriteString("### Disks mounted/unmounted (derived from lsblk -J)\n\n```\n" + safe(rep.Summary.LsblkMountView) + "\n```\n\n")
	b.WriteString("### Mounts\n\n```\n" + safe(rep.Summary.Mounts) + "\n```\n\n")
	b.WriteString("### Network\n\n```\n" + safe(rep.Summary.NetOverview) + "\n```\n\n")
	b.WriteString("### Failed systemd units\n\n```\n" + safe(rep.Summary.FailedSystemd) + "\n```\n\n")
	b.WriteString("### Recent reboots/logins (last -x)\n\n```\n" + safe(rep.Summary.RecentReboots) + "\n```\n\n")
	b.WriteString("### SMART\n\n```\n" + safe(rep.Summary.SmartctlInfo) + "\n```\n\n")

	b.WriteString("### OOM/Kernel hints\n\n")
	b.WriteString("- " + safe(rep.Summary.OOMKills) + "\n")
	b.WriteString("- " + safe(rep.Summary.DmesgErrorsBrief) + "\n\n")

	b.WriteString("### Top CPU\n\n```\n" + safe(rep.Summary.TopCPU) + "\n```\n\n")
	b.WriteString("### Top MEM\n\n```\n" + safe(rep.Summary.TopMEM) + "\n```\n\n")

	b.WriteString("## Findings (errors/failures)\n\n")
	b.WriteString(fmt.Sprintf("Total findings: **%d** (deduped/capped)\n\n", len(rep.Findings)))

	maxRows := 250
	b.WriteString("| # | Source | Severity | Message |\n")
	b.WriteString("|---:|---|---|---|\n")
	for i, f := range rep.Findings {
		if i >= maxRows {
			b.WriteString("| … | … | … | (see errors.csv for full list) |\n")
			break
		}
		msg := escapeMD(f.Message)
		if len(msg) > 350 {
			msg = msg[:350] + "…"
		}
		b.WriteString(fmt.Sprintf("| %d | `%s` | `%s` | %s |\n", i+1, escapeMD(f.Source), escapeMD(f.Severity), msg))
	}

	if len(rep.FileStats) > 0 {
		b.WriteString("\n## Log scan stats\n\n```\n")
		for _, s := range rep.FileStats {
			b.WriteString(s + "\n")
		}
		b.WriteString("```\n")
	}

	must(os.WriteFile(path, []byte(b.String()), 0o644))
}

func safe(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(no data / command failed / permission denied)"
	}
	return s
}

func escapeMD(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}
