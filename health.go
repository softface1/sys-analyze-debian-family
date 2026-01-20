

package main

import (
	"fmt"
	"strconv"
	"strings"
)

func computeHealth(failedSystemd string, df string, findings []Finding, oomCount int, unmountedFS int, smartBad int) Health {
	h := Health{Score: 100}

	// Failed units
	h.FailedUnits = countFailedUnits(failedSystemd)
	if h.FailedUnits > 0 {
		pen := minInt(35, 5*h.FailedUnits)
		h.Score -= pen
		h.Reasons = append(h.Reasons, fmt.Sprintf("systemd failed units: %d (-%d)", h.FailedUnits, pen))
	}

	// Disk critical mounts >= 90%
	h.DiskCritical = countDiskCritical(df)
	if h.DiskCritical > 0 {
		pen := minInt(25, 6*h.DiskCritical)
		h.Score -= pen
		h.Reasons = append(h.Reasons, fmt.Sprintf("disk usage >= 90%% mounts: %d (-%d)", h.DiskCritical, pen))
	}

	// OOM
	h.OOMCount = oomCount
	if h.OOMCount > 0 {
		pen := minInt(30, 10*h.OOMCount)
		h.Score -= pen
		h.Reasons = append(h.Reasons, fmt.Sprintf("OOM keywords: %d (-%d)", h.OOMCount, pen))
	}

	// I/O 
	h.IOErrors = countFindingsByKeywords(findings, []string{
		"i/o error", "buffer i/o error", "blk_update_request", "ata", "nvme", "reset",
		"no route to host", "network is unreachable", "connection refused",
		"timed out", "timeout",
	})
	if h.IOErrors > 0 {
		pen := minInt(25, 2*h.IOErrors)
		h.Score -= pen
		h.Reasons = append(h.Reasons, fmt.Sprintf("I/O/network-ish error hits: %d (-%d)", h.IOErrors, pen))
	}

	// FS errors 
	h.FSErrors = countFindingsByKeywords(findings, []string{
		"ext4-fs error", "xfs", "corrupt", "superblock", "metadata", "read-only file system",
	})
	if h.FSErrors > 0 {
		pen := minInt(25, 3*h.FSErrors)
		h.Score -= pen
		h.Reasons = append(h.Reasons, fmt.Sprintf("filesystem error hits: %d (-%d)", h.FSErrors, pen))
	}

	// Unmounted fs part
	h.UnmountedFSTypes = unmountedFS
	if unmountedFS > 0 {
		h.Reasons = append(h.Reasons, fmt.Sprintf("unmounted filesystem partitions detected (excluding swap): %d", unmountedFS))
	} else {
		h.Reasons = append(h.Reasons, "no unmounted filesystem partitions detected (excluding swap)")
	}

	// SMART bad 
	h.SmartBadHealth = smartBad
	if smartBad > 0 {
		pen := minInt(40, 20*smartBad)
		h.Score -= pen
		h.Reasons = append(h.Reasons, fmt.Sprintf("SMART overall-health FAILED disks: %d (-%d)", smartBad, pen))
	}

	if h.Score < 0 {
		h.Score = 0
	}
	h.Grade = grade(h.Score)

	if len(h.Reasons) == 0 {
		h.Reasons = []string{"no obvious red flags detected"}
	}
	return h
}

func grade(score int) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 75:
		return "B"
	case score >= 60:
		return "C"
	case score >= 40:
		return "D"
	default:
		return "F"
	}
}

func countFailedUnits(out string) int {
	out = strings.TrimSpace(out)
	if out == "" || strings.Contains(out, "0 loaded units listed") {
		return 0
	}

	n := 0
	for _, ln := range strings.Split(out, "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		if strings.Contains(ln, ".service") || strings.Contains(ln, ".mount") || strings.Contains(ln, ".socket") {
			if strings.HasPrefix(ln, "UNIT ") {
				continue
			}
			n++
		}
	}
	if n == 0 {
		for _, ln := range strings.Split(out, "\n") {
			if strings.Contains(strings.ToLower(ln), "failed") {
				n++
			}
		}
	}
	return n
}

func countDiskCritical(dfOut string) int {
	dfOut = strings.TrimSpace(dfOut)
	if dfOut == "" {
		return 0
	}
	lines := strings.Split(dfOut, "\n")
	if len(lines) <= 1 {
		return 0
	}

	crit := 0
	for _, ln := range lines[1:] {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		// FS TYPE SIZE USED AVAIL USE% MOUNT
		fields := strings.Fields(ln)
		if len(fields) < 7 {
			continue
		}
		use := strings.TrimSuffix(fields[5], "%")
		p, err := strconv.Atoi(use)
		if err != nil {
			continue
		}
		if p >= 90 {
			crit++
		}
	}
	return crit
}

func countFindingsByKeywords(findings []Finding, kws []string) int {
	lowK := make([]string, 0, len(kws))
	for _, k := range kws {
		lowK = append(lowK, strings.ToLower(k))
	}

	allowSev := map[string]bool{
		"ERR":      true,
		"WARN":     true,
		"WARN/ERR": true,
	}

	n := 0
	for _, f := range findings {
		if !allowSev[f.Severity] {
			continue
		}
		s := strings.ToLower(f.Message)
		for _, k := range lowK {
			if strings.Contains(s, k) {
				n++
				break
			}
		}
	}
	return n
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
