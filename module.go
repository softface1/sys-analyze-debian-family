package main

type Module interface {
	ID() string
	Run(e *Exec, rep *Report, cfg Config) error
}

func defaultModules() []Module {
//add new module here

	return []Module{
		ModBase{},
		ModDisk{},  
		ModSmart{}, 
		ModNet{},
		ModSystemd{},
		ModLogs{},   
		ModHealth{}, 
	}
}
