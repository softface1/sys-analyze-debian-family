package main

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func packDirTarGz(dir string, outFile string) error {
	outAbs, _ := filepath.Abs(outFile)

	f, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	baseAbs, _ := filepath.Abs(dir)

	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		abs, _ := filepath.Abs(path)
		if abs == outAbs {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		rel, err := filepath.Rel(baseAbs, abs)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}

		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return nil
		}
		hdr.Name = rel
		if info.IsDir() {
			hdr.Name += "/"
		}

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		if info.Mode().IsRegular() {
			rf, err := os.Open(path)
			if err != nil {
				return nil
			}
			defer rf.Close()
			_, _ = io.Copy(tw, rf)
		}
		return nil
	})
}
