package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type ModLogs struct{}

func (ModLogs) ID() string { return "logs" }

func (ModLogs) Run(e *Exec, rep *Report, cfg Config) error {
	jErr := strings.TrimSpace(e.Run(`journalctl --since "-` + cfg.Since + `" -p err..alert --no-pager -n 1500`).Stdout)
	rep.Findings = append(rep.Findings, parseJournalFindings(jErr)...)

	kErr := strings.TrimSpace(e.Run(`journalctl -k -b -p err..alert --no-pager -n 300`).Stdout)
	if kErr != "" {
		rep.Findings = append(rep.Findings, parseJournalFindings(kErr)...)
	}

	dmesg := strings.TrimSpace(e.Run(`dmesg --level=err,warn 2>/dev/null | tail -n 600 || dmesg -T | tail -n 600`).Stdout)
	dmesgFind, dmesgBrief := parseDmesgFindings(dmesg)
	rep.Findings = append(rep.Findings, dmesgFind...)
	rep.Summary.DmesgErrorsBrief = dmesgBrief

	oomCount := countKeywordHits(appendLines(jErr, dmesg), []string{
		"Out of memory", "oom-kill", "Killed process", "invoked oom-killer",
	})
	rep.Summary.OOMKills = fmt.Sprintf("OOM hits (keywords): %d", oomCount)
	rep.Health.OOMCount = oomCount

	logFiles := discoverLogFilesAstra()
	rep.LogFiles = logFiles

	for _, lf := range logFiles {
		ff, stats := scanLogFile(lf)
		if stats != "" {
			rep.FileStats = append(rep.FileStats, stats)
		}
		rep.Findings = append(rep.Findings, ff...)
	}

	return nil
}

func parseJournalFindings(j string) []Finding {
	j = strings.TrimSpace(j)
	if j == "" {
		return nil
	}
	lines := strings.Split(j, "\n")
	out := make([]Finding, 0, len(lines))
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		out = append(out, Finding{
			Source:   "journalctl",
			Severity: "ERR",
			Message:  ln,
		})
	}
	return out
}

func parseDmesgFindings(d string) ([]Finding, string) {
	d = strings.TrimSpace(d)
	if d == "" {
		return nil, ""
	}

	keywords := []string{
		"error", "failed", "fail", "critical", "panic", "segfault",
		"I/O error", "EXT4-fs error", "XFS", "Buffer I/O error",
		"blk_update_request", "ata", "nvme", "reset", "hung task",
		"Out of memory", "oom-kill", "Killed process",
	}

	re := regexp.MustCompile(`(?i)` + strings.Join(escapeKeywords(keywords), "|"))
	lines := strings.Split(d, "\n")
	find := make([]Finding, 0, 128)
	hits := 0
	for _, ln := range lines {
		if re.MatchString(ln) {
			hits++
			find = append(find, Finding{
				Source:   "dmesg",
				Severity: "WARN/ERR",
				Message:  strings.TrimSpace(ln),
			})
		}
	}
	brief := fmt.Sprintf("dmesg keyword hits: %d (scanned last ~600 lines)", hits)
	return find, brief
}

func escapeKeywords(kw []string) []string {
	out := make([]string, 0, len(kw))
	for _, k := range kw {
		out = append(out, regexp.QuoteMeta(k))
	}
	return out
}

func appendLines(a, b string) string {
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	return a + "\n" + b
}

func countKeywordHits(text string, patterns []string) int {
	t := strings.ToLower(text)
	total := 0
	for _, p := range patterns {
		total += strings.Count(t, strings.ToLower(p))
	}
	return total
}

func discoverLogFilesAstra() []string {
	var files []string

	addIfExists := func(p string) {
		if fileExists(p) {
			files = append(files, p)
		}
	}

	common := []string{
		"/var/log/syslog",
		"/var/log/messages",
		"/var/log/kern.log",
		"/var/log/auth.log",
		"/var/log/daemon.log",
		"/var/log/user.log",
		"/var/log/dpkg.log",
		"/var/log/apt/history.log",
		"/var/log/audit/audit.log",
		"/var/log/boot.log",
	}
	for _, p := range common {
		addIfExists(p)
	}

	roots := []string{"/var/log", "/var/log/syslog-ng", "/var/log/audit"}
	for _, root := range roots {
		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			rel, _ := filepath.Rel(root, path)
			if strings.Count(rel, string(os.PathSeparator)) > 3 {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if d.IsDir() {
				return nil
			}
			low := strings.ToLower(path)
			if strings.HasSuffix(low, ".log") || strings.HasSuffix(low, ".log.1") || strings.Contains(low, "audit.log") {
				files = append(files, path)
			}
			return nil
		})
	}

	m := map[string]bool{}
	var uniq []string
	for _, f := range files {
		if !m[f] {
			m[f] = true
			uniq = append(uniq, f)
		}
	}
	sort.Strings(uniq)
	return uniq
}

func scanLogFile(path string) ([]Finding, string) {
	st, err := os.Stat(path)
	if err != nil || st.IsDir() {
		return nil, ""
	}

	const maxBytes = 5_000_000
	f, err := os.Open(path)
	if err != nil {
		return []Finding{{
			Source:   path,
			Severity: "INFO",
			Message:  "cannot open log: " + err.Error(),
		}}, fmt.Sprintf("%s: open failed", path)
	}
	defer f.Close()

	if st.Size() > maxBytes {
		_, _ = f.Seek(st.Size()-maxBytes, io.SeekStart)
	}

	re := regexp.MustCompile(`(?i)\b(err(or)?|failed|failure|panic|critical|segfault|fatal|i/o error|corrupt|denied|unauthorized|timeout|timed out|refused|unreachable|oom|killed process)\b`)

	sc := bufio.NewScanner(f)
	buf := make([]byte, 0, 128*1024)
	sc.Buffer(buf, 2*1024*1024)

	find := make([]Finding, 0, 128)
	linesScanned := 0
	hits := 0
	for sc.Scan() {
		linesScanned++
		ln := sc.Text()
		if re.MatchString(ln) {
			hits++
			msg := strings.TrimSpace(ln)
			if len(msg) > 2000 {
				msg = msg[:2000] + "...[cut]"
			}
			find = append(find, Finding{
				Source:   path,
				Severity: "MATCH",
				Message:  msg,
			})
			if len(find) > 900 {
				break
			}
		}
	}

	stats := fmt.Sprintf("%s: size=%dB scanned_lines~%d hits=%d", path, st.Size(), linesScanned, hits)
	return find, stats
}
