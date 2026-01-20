package main

type ModHealth struct{}

func (m ModHealth) ID() string { return "health" }

func (m ModHealth) Run(e *Exec, rep *Report, cfg Config) error {

	rep.Health = computeHealth(
		rep.Summary.FailedSystemd,
		rep.Summary.DiskDf,
		rep.Findings,
		rep.Health.OOMCount,
		rep.Health.UnmountedFSTypes,
		rep.Health.SmartBadHealth,
	)

	return nil
}
