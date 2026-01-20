package main

import "strings"

type ModBase struct{}

func (ModBase) ID() string { return "base" }

func (ModBase) Run(e *Exec, rep *Report, cfg Config) error {
	e.Run("date -Is")
	rep.Summary.Uptime = strings.TrimSpace(e.Run("uptime -p").Stdout)
	rep.Summary.LoadAvg = strings.TrimSpace(e.Run("cat /proc/loadavg").Stdout)
	rep.Summary.MemInfo = strings.TrimSpace(e.Run("free -h").Stdout)

	rep.Summary.TopCPU = strings.TrimSpace(e.Run("ps -eo pid,ppid,user,comm,%cpu,%mem,etime --sort=-%cpu | head -n 25").Stdout)
	rep.Summary.TopMEM = strings.TrimSpace(e.Run("ps -eo pid,ppid,user,comm,%cpu,%mem,etime --sort=-%mem | head -n 25").Stdout)

	rep.Summary.RecentReboots = strings.TrimSpace(e.Run("last -x | head -n 20").Stdout)
	return nil
}
