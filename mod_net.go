package main

import "strings"

type ModNet struct{}

func (ModNet) ID() string { return "net" }

func (ModNet) Run(e *Exec, rep *Report, cfg Config) error {
	ipa := strings.TrimSpace(e.Run("ip -br a").Stdout)
	ipr := strings.TrimSpace(e.Run("ip r").Stdout)
	ss := strings.TrimSpace(e.Run("ss -tulpen").Stdout)
	sss := strings.TrimSpace(e.Run("ss -s").Stdout)

	e.Run("ip -s link")

	rep.Summary.NetOverview = strings.TrimSpace(ipa + "\n\n" + ipr + "\n\n" + ss + "\n\n" + sss)
	return nil
}
