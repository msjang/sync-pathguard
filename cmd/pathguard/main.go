// Command pathguard is the headless CLI: it scans the configured (or given)
// folders and prints a report, or a summary JSON with --json. Pure Go (no cgo),
// so it cross-compiles freely and is handy for scripts and CI. Exits non-zero
// when any name/path is over the limit.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/msjang/sync-pathguard/internal/config"
	"github.com/msjang/sync-pathguard/internal/scan"
)

func main() {
	jsonOut := flag.Bool("json", false, "print a summary as JSON")
	cfgPath := flag.String("config", "", "config file path (default: OS config dir)")
	root := flag.String("root", "", "scan this folder instead of the configured watches")
	remote := flag.String("remote-prefix", "", "remote absolute root for PATH_MAX (used with --root)")
	flag.Parse()

	cfg := config.Default()
	switch {
	case *root != "":
		rp := *remote
		if rp == "" {
			rp = config.Default().Watch[0].RemotePrefix
		}
		cfg.Watch = []config.Watch{{Root: *root, RemotePrefix: rp}}
	case *cfgPath != "":
		c, err := config.LoadFile(*cfgPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "config:", err)
			os.Exit(2)
		}
		cfg = c
	default:
		c, _, err := config.Load()
		if err != nil {
			fmt.Fprintln(os.Stderr, "config:", err, "(using defaults)")
		} else {
			cfg = c
		}
	}

	lim := scan.Limits{NameMax: cfg.Limits.NameMax, PathMax: cfg.Limits.PathMax, WarnRatio: cfg.Limits.WarnRatio}
	excl := cfg.ExcludeSet()
	var agg scan.Result
	for _, w := range cfg.Watch {
		res, err := scan.Scan(config.ExpandRoot(w.Root), w.RemotePrefix, lim, excl)
		if err != nil {
			fmt.Fprintf(os.Stderr, "scan %s: %v\n", w.Root, err)
		}
		agg = merge(agg, res)
	}

	if *jsonOut {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]int{
			"total":     agg.Total,
			"name_over": len(agg.NameOver), "name_warn": len(agg.NameWarn),
			"path_over": len(agg.PathOver), "path_warn": len(agg.PathWarn),
		})
	} else {
		printReport(agg, lim)
	}

	if len(agg.NameOver)+len(agg.PathOver) > 0 {
		os.Exit(1)
	}
}

func printReport(r scan.Result, lim scan.Limits) {
	fmt.Printf("Scanned %d entries (NAME_MAX=%dB, PATH_MAX=%dB, NFD worst-case)\n", r.Total, lim.NameMax, lim.PathMax)
	fmt.Printf("  names  over(>%dB): %d | warn: %d\n", lim.NameMax, len(r.NameOver), len(r.NameWarn))
	fmt.Printf("  paths  over(>%dB): %d | warn: %d\n", lim.PathMax, len(r.PathOver), len(r.PathWarn))

	show := func(title string, rs []scan.Record, n int) {
		if len(rs) == 0 {
			return
		}
		fmt.Printf("\n■ %s (%d):\n", title, len(rs))
		if n > len(rs) {
			n = len(rs)
		}
		for _, rec := range rs[:n] {
			fmt.Printf("    NFD %3dB (cur %dB/%s, NFC %dB)  %s\n",
				rec.NameNFD, rec.NameCur, rec.Form, rec.NameNFC, rec.Rel)
		}
	}
	show("NAME_MAX over", r.NameOver, 40)
	show("NAME_MAX warn", r.NameWarn, 15)
	if len(r.PathOver) > 0 {
		fmt.Printf("\n■ PATH_MAX over (%d):\n", len(r.PathOver))
		for _, rec := range r.PathOver[:min(20, len(r.PathOver))] {
			fmt.Printf("    NFD %dB  %s\n", rec.PathNFD, rec.Rel)
		}
	}
}

func merge(a, b scan.Result) scan.Result {
	a.Total += b.Total
	a.NameOver = append(a.NameOver, b.NameOver...)
	a.NameWarn = append(a.NameWarn, b.NameWarn...)
	a.PathOver = append(a.PathOver, b.PathOver...)
	a.PathWarn = append(a.PathWarn, b.PathWarn...)
	return a
}
