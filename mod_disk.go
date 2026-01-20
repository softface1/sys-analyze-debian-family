package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ModDisk struct{}

func (ModDisk) ID() string { return "disk" }

func (ModDisk) Run(e *Exec, rep *Report, cfg Config) error {
	rep.Summary.DiskDf = strings.TrimSpace(e.Run("df -hT").Stdout)
	e.Run("df -i")

	rep.Summary.LsblkText = strings.TrimSpace(e.Run("lsblk -o NAME,TYPE,SIZE,FSTYPE,MOUNTPOINT,UUID,MODEL,SERIAL").Stdout)
	lsblkJSONOut := strings.TrimSpace(e.Run("lsblk -J -o NAME,TYPE,SIZE,FSTYPE,MOUNTPOINT,UUID,MODEL,SERIAL").Stdout)

	rep.Summary.Mounts = strings.TrimSpace(e.Run("mount").Stdout)

	mountView, unmountedFindings, unmountedCount, disks := analyzeLsblk(lsblkJSONOut)
	rep.Summary.LsblkMountView = mountView
	rep.Findings = append(rep.Findings, unmountedFindings...)
	rep.Health.UnmountedFSTypes = unmountedCount
	rep.Disks = disks

	return nil
}

func analyzeLsblk(lsblkJSONOut string) (mountView string, findings []Finding, unmountedCount int, disks []DiskRef) {
	lsblkJSONOut = strings.TrimSpace(lsblkJSONOut)
	if lsblkJSONOut == "" {
		return "(no lsblk -J data)", nil, 0, nil
	}

	var j LsblkJSON
	if err := json.Unmarshal([]byte(lsblkJSONOut), &j); err != nil {
		return "lsblk -J parse failed: " + err.Error(), []Finding{{
			Source:   "lsblk",
			Severity: "INFO",
			Message:  "lsblk -J parse failed: " + err.Error(),
		}}, 0, nil
	}

	var b strings.Builder
	b.WriteString("Legend: [MOUNTED] vs [UNMOUNTED]\n\n")

	var warns []Finding
	unmounted := 0
	var diskRefs []DiskRef

	var walk func(dev LsblkDev, indent string)
	walk = func(dev LsblkDev, indent string) {
		name := dev.Name
		if name != "" && !strings.HasPrefix(name, "/dev/") {
			name = "/dev/" + name
		}

		mp := strings.TrimSpace(dev.Mountpoint)
		fsType := strings.TrimSpace(dev.Fstype)

		tag := "[UNMOUNTED]"
		if mp != "" {
			tag = "[MOUNTED]"
		}

		line := fmt.Sprintf("%s%s %s type=%s size=%s", indent, tag, name, dev.Type, dev.Size)
		if fsType != "" {
			line += " fstype=" + fsType
		}
		if mp != "" {
			line += " mount=" + mp
		}
		if dev.Type == "disk" {
			ms := strings.TrimSpace(strings.Join([]string{strings.TrimSpace(dev.Model), strings.TrimSpace(dev.Serial)}, " "))
			if ms != "" {
				line += " (" + ms + ")"
			}
		}
		b.WriteString(line + "\n")

		if mp == "" && fsType != "" && strings.ToLower(fsType) != "swap" {
			if dev.Type == "part" || dev.Type == "lvm" {
				unmounted++
				warns = append(warns, Finding{
					Source:   "lsblk",
					Severity: "WARN",
					Message:  fmt.Sprintf("%s has fstype=%s but is NOT mounted", name, fsType),
				})
			}
		}

		for _, ch := range dev.Children {
			walk(ch, indent+"  ")
		}
	}

	for _, d := range j.Blockdevices {
		if d.Type == "loop" || d.Type == "rom" {
			continue
		}
		if d.Type == "disk" {
			name := d.Name
			if name != "" && !strings.HasPrefix(name, "/dev/") {
				name = "/dev/" + name
			}
			diskRefs = append(diskRefs, DiskRef{
				Name:   name,
				Model:  strings.TrimSpace(d.Model),
				Serial: strings.TrimSpace(d.Serial),
				Size:   strings.TrimSpace(d.Size),
			})
		}
		walk(d, "")
		b.WriteString("\n")
	}

	if len(warns) == 0 {
		warns = append(warns, Finding{
			Source:   "lsblk",
			Severity: "INFO",
			Message:  "no unmounted filesystem partitions detected (excluding swap)",
		})
	}

	return strings.TrimSpace(b.String()), warns, unmounted, diskRefs
}
