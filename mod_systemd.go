package main

import "strings"

type ModSystemd struct{}

func (ModSystemd) ID() string { return "systemd" }

func (ModSystemd) Run(e *Exec, rep *Report, cfg Config) error {
	rep.Summary.FailedSystemd = strings.TrimSpace(e.Run("systemctl --failed --no-pager").Stdout)
	e.Run("systemctl list-units --state=failed --no-legend --plain")
	return nil
}
