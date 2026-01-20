package main

import "time"

type Config struct {
	OutDir      string
	Since       string
	KeepRaw     bool
	MaxFindings int
	CmdTimeout  time.Duration
	MakeHTML    bool
	DoPack      bool
}
