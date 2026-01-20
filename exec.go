package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Exec struct {
	Timeout  time.Duration
	KeepRaw  bool
	RawDir   string
	Commands *[]CommandOut
}

func (e *Exec) Run(cmdline string) CommandOut {
	out := runCmd(e.Timeout, cmdline)
	if e.Commands != nil {
		*e.Commands = append(*e.Commands, out)
	}
	if e.KeepRaw && e.RawDir != "" {
		saveRaw(e.RawDir, cmdline, out)
	}
	return out
}

func (e *Exec) Has(bin string) bool {
	out := e.Run("command -v " + bin + " >/dev/null 2>&1; echo $?")
	return strings.TrimSpace(out.Stdout) == "0"
}

func runCmd(timeout time.Duration, cmdline string) CommandOut {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c := exec.CommandContext(ctx, "bash", "-lc", cmdline)
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()
	exit := 0
	if err != nil {
		exit = exitCode(err)
	}
	if ctx.Err() == context.DeadlineExceeded {
		stderr.WriteString("\n[timeout exceeded]\n")
		exit = 124
	}

	return CommandOut{
		Cmd:      cmdline,
		ExitCode: exit,
		Stdout:   trimHuge(stdout.String(), 300_000),
		Stderr:   trimHuge(stderr.String(), 120_000),
	}
}

func exitCode(err error) int {
	type exitCoder interface{ ExitCode() int }
	if ee, ok := err.(exitCoder); ok {
		return ee.ExitCode()
	}
	return 1
}

func trimHuge(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "\n...[truncated]...\n"
}

func saveRaw(rawDir, cmd string, out CommandOut) {
	sum := sha256.Sum256([]byte(cmd))
	name := fmt.Sprintf("%x.txt", sum[:8])
	path := filepath.Join(rawDir, name)

	var b strings.Builder
	b.WriteString("CMD: " + cmd + "\n")
	b.WriteString("EXIT: " + fmt.Sprint(out.ExitCode) + "\n\n")
	if out.Stdout != "" {
		b.WriteString("STDOUT:\n" + out.Stdout + "\n")
	}
	if out.Stderr != "" {
		b.WriteString("\nSTDERR:\n" + out.Stderr + "\n")
	}
	_ = os.WriteFile(path, []byte(b.String()), 0o644)
}
