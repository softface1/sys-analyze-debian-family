package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type ModSmart struct{}

func (ModSmart) ID() string { return "smart" }

func (ModSmart) Run(e *Exec, rep *Report, cfg Config) error {
	smartInfo, smartFindings, smartBad := collectSMART(e, 12*time.Second, rep.Disks)
	rep.Summary.SmartctlInfo = smartInfo
	rep.Findings = append(rep.Findings, smartFindings...)
	rep.Health.SmartBadHealth = smartBad
	return nil
}

func collectSMART(e *Exec, timeout time.Duration, disks []DiskRef) (summary string, findings []Finding, smartBad int) {
	if len(disks) == 0 {
		return "SMART: no disks detected (lsblk)", []Finding{{
			Source:   "smartctl",
			Severity: "INFO",
			Message:  "no disks detected via lsblk, SMART skipped",
		}}, 0
	}

	if !e.Has("smartctl") {
		msg := "SMART: smartctl is not installed (install smartmontools). Skipping SMART checks."
		return msg, []Finding{{
			Source:   "smartctl",
			Severity: "INFO",
			Message:  msg,
		}}, 0
	}

	var b strings.Builder
	b.WriteString("SMART (smartctl -H -A) summary\n\n")

	var fs []Finding
	bad := 0

	reOverall := regexp.MustCompile(`(?mi)^\s*SMART overall-health self-assessment test result:\s*(PASSED|FAILED)\s*$`)
	reHealth := regexp.MustCompile(`(?mi)^\s*SMART Health Status:\s*(OK|PASSED|FAILED)\s*$`)

	oldTimeout := e.Timeout
	e.Timeout = timeout
	defer func() { e.Timeout = oldTimeout }()

	for _, d := range disks {
		dev := strings.TrimSpace(d.Name)
		if dev == "" || !strings.HasPrefix(dev, "/dev/") {
			continue
		}

		cmd := fmt.Sprintf("smartctl -n standby -H -A %s 2>&1", shellEscape(dev))
		out := e.Run(cmd)
		text := strings.TrimSpace(out.Stdout + "\n" + out.Stderr)
		if text == "" {
			text = "(no output)"
		}

		header := fmt.Sprintf("=== %s %s %s %s ===\n", dev, nz(d.Size), nz(d.Model), nz(d.Serial))
		b.WriteString(header)
		b.WriteString(text + "\n\n")

		low := strings.ToLower(text)

		health := ""
		if m := reOverall.FindStringSubmatch(text); len(m) == 2 {
			health = strings.ToUpper(m[1]) // PASSED/FAILED
		} else if m := reHealth.FindStringSubmatch(text); len(m) == 2 {
			health = strings.ToUpper(m[1]) // OK/PASSED/FAILED
			if health == "OK" {
				health = "PASSED"
			}
		} else {
			fs = append(fs, Finding{
				Source:   "smartctl",
				Severity: "INFO",
				Message:  fmt.Sprintf("%s SMART health line not found (device may not support it)", dev),
			})
		}

		if health == "FAILED" {
			bad++
			fs = append(fs, Finding{
				Source:   "smartctl",
				Severity: "ERR",
				Message:  fmt.Sprintf("%s SMART overall-health reports FAILED", dev),
			})
		}

		if strings.Contains(low, "reallocated_sector") || strings.Contains(low, "reallocated sector") {
			fs = append(fs, Finding{
				Source:   "smartctl",
				Severity: "WARN",
				Message:  fmt.Sprintf("%s has Reallocated_Sector related attribute (check RAW_VALUE)", dev),
			})
		}
		if strings.Contains(low, "current_pending_sector") || (strings.Contains(low, "pending") && strings.Contains(low, "sector")) {
			fs = append(fs, Finding{
				Source:   "smartctl",
				Severity: "WARN",
				Message:  fmt.Sprintf("%s shows pending sectors (check RAW_VALUE)", dev),
			})
		}
		if strings.Contains(low, "offline_uncorrectable") || (strings.Contains(low, "uncorrect") && strings.Contains(low, "error")) {
			fs = append(fs, Finding{
				Source:   "smartctl",
				Severity: "WARN",
				Message:  fmt.Sprintf("%s shows uncorrectable errors (check RAW_VALUE)", dev),
			})
		}
	}

	if len(fs) == 0 {
		fs = append(fs, Finding{
			Source:   "smartctl",
			Severity: "INFO",
			Message:  "SMART collected; no FAILED statuses detected",
		})
	}

	return strings.TrimSpace(b.String()), fs, bad
}
