package main

type Finding struct {
	Source   string `json:"source"`
	When     string `json:"when,omitempty"`
	Severity string `json:"severity,omitempty"`
	Message  string `json:"message"`
	Context  string `json:"context,omitempty"`
}

type CommandOut struct {
	Cmd      string `json:"cmd"`
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

type Health struct {
	Score        int      `json:"score"`
	Grade        string   `json:"grade"`
	Reasons      []string `json:"reasons"`
	FailedUnits  int      `json:"failed_units"`
	OOMCount     int      `json:"oom_count"`
	IOErrors     int      `json:"io_errors"`
	FSErrors     int      `json:"fs_errors"`
	DiskCritical int      `json:"disk_critical_mounts"` 

	UnmountedFSTypes int `json:"unmounted_fs_types"` 
	SmartBadHealth   int `json:"smart_bad_health"`   
}

type Report struct {
	Meta struct {
		GeneratedAt string `json:"generated_at"`
		Hostname    string `json:"hostname"`
		Kernel      string `json:"kernel"`
		OS          string `json:"os"`
		Arch        string `json:"arch"`
		User        string `json:"user"`
		Since       string `json:"since"`
	} `json:"meta"`

	Summary struct {
		Uptime           string `json:"uptime,omitempty"`
		LoadAvg          string `json:"loadavg,omitempty"`
		MemInfo          string `json:"meminfo,omitempty"`
		DiskDf           string `json:"disk_df,omitempty"`
		TopCPU           string `json:"top_cpu,omitempty"`
		TopMEM           string `json:"top_mem,omitempty"`
		NetOverview      string `json:"net_overview,omitempty"`
		FailedSystemd    string `json:"failed_systemd,omitempty"`
		RecentReboots    string `json:"recent_reboots,omitempty"`
		OOMKills         string `json:"oom_kills,omitempty"`
		DmesgErrorsBrief string `json:"dmesg_errors_brief,omitempty"`

		LsblkText      string `json:"lsblk_text,omitempty"`
		LsblkMountView string `json:"lsblk_mount_view,omitempty"`
		Mounts         string `json:"mounts,omitempty"`

		SmartctlInfo string `json:"smartctl_info,omitempty"`
	} `json:"summary"`

	// NEW: typed disks list
	Disks []DiskRef `json:"disks,omitempty"`

	Health    Health         `json:"health"`
	Commands  []CommandOut   `json:"commands"`
	Findings  []Finding      `json:"findings"`
	LogFiles  []string       `json:"log_files"`
	FileStats []string       `json:"file_stats,omitempty"`
	Extra     map[string]any `json:"extra,omitempty"` 
}

type LsblkJSON struct {
	Blockdevices []LsblkDev `json:"blockdevices"`
}

type LsblkDev struct {
	Name       string     `json:"name"`
	Type       string     `json:"type"`
	Size       string     `json:"size"`
	Fstype     string     `json:"fstype"`
	Mountpoint string     `json:"mountpoint"`
	UUID       string     `json:"uuid"`
	Model      string     `json:"model"`
	Serial     string     `json:"serial"`
	Children   []LsblkDev `json:"children"`
}


type DiskRef struct {
	Name   string `json:"name"`   
	Model  string `json:"model"`  
	Serial string `json:"serial"` 
	Size   string `json:"size"`
}
