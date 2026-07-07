// Package scan implements the read-only NFD byte-length check that is the core
// of Pathguard. It mirrors the reference pathguard.py: every file/dir name
// and full (remote) path is measured as UTF-8 bytes after NFD normalization —
// the worst case for combining-mark scripts like Korean.
package scan

import (
	"io/fs"
	"path/filepath"
	"sort"

	"golang.org/x/text/unicode/norm"
)

// Record is one file/dir entry with its measured byte lengths.
type Record struct {
	Rel     string // path relative to root, forward-slashed
	Abs     string // absolute local path (for reveal in file manager)
	Form    string // "NFC" | "NFD" | "mixed"
	NameCur int    // current on-disk name byte length
	NameNFC int    // name length if normalized to NFC
	NameNFD int    // name length if normalized to NFD (worst case)
	PathNFD int    // full remote path length in NFD (worst case)
}

// Limits are the byte thresholds to check against.
type Limits struct {
	NameMax   int
	PathMax   int
	WarnRatio float64 // e.g. 0.80 → warn from 80% of the limit
}

// Result groups the findings of a scan.
type Result struct {
	Total    int
	NameOver []Record
	NameWarn []Record
	PathOver []Record
	PathWarn []Record
}

func bNFD(s string) int { return len(norm.NFD.String(s)) }
func bNFC(s string) int { return len(norm.NFC.String(s)) }

func formOf(s string) string {
	switch {
	case s == norm.NFC.String(s):
		return "NFC"
	case s == norm.NFD.String(s):
		return "NFD"
	default:
		return "mixed"
	}
}

// Scan walks root and returns every name/path that exceeds (over) or approaches
// (warn) the limits. Names in exclude are skipped; excluded directories are not
// descended into. remotePrefix is the destination absolute root used for the
// PATH_MAX calculation. I/O errors on individual entries are ignored so one
// unreadable directory does not abort the whole scan.
func Scan(root, remotePrefix string, lim Limits, exclude map[string]bool) (Result, error) {
	var res Result
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if path == root {
			return nil
		}
		name := d.Name()
		if exclude[name] {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		res.Total++
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		nameNFD := bNFD(name)
		pathNFD := bNFD(remotePrefix + "/" + rel)
		rec := Record{
			Rel: rel, Abs: path, Form: formOf(name),
			NameCur: len(name), NameNFC: bNFC(name), NameNFD: nameNFD,
			PathNFD: pathNFD,
		}
		if nameNFD > lim.NameMax {
			res.NameOver = append(res.NameOver, rec)
		} else if float64(nameNFD) >= float64(lim.NameMax)*lim.WarnRatio {
			res.NameWarn = append(res.NameWarn, rec)
		}
		if pathNFD > lim.PathMax {
			res.PathOver = append(res.PathOver, rec)
		} else if float64(pathNFD) >= float64(lim.PathMax)*lim.WarnRatio {
			res.PathWarn = append(res.PathWarn, rec)
		}
		return nil
	})

	// Worst first, so the tray menu and report show the biggest offenders on top.
	byNameNFD := func(rs []Record) { sort.SliceStable(rs, func(i, j int) bool { return rs[i].NameNFD > rs[j].NameNFD }) }
	byPathNFD := func(rs []Record) { sort.SliceStable(rs, func(i, j int) bool { return rs[i].PathNFD > rs[j].PathNFD }) }
	byNameNFD(res.NameOver)
	byNameNFD(res.NameWarn)
	byPathNFD(res.PathOver)
	byPathNFD(res.PathWarn)
	return res, err
}
