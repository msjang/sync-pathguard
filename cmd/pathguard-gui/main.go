// Command pathguard-gui is the resident tray / menu-bar app.
// It periodically (and on demand) scans the configured folders for names/paths
// whose NFD byte length risks breaking cloud/NAS sync, reflects the result in
// the tray icon color, and lets you jump straight to an offending file.
package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"

	"fyne.io/systray"

	"github.com/msjang/pathguard/internal/config"
	"github.com/msjang/pathguard/internal/i18n"
	"github.com/msjang/pathguard/internal/report"
	"github.com/msjang/pathguard/internal/scan"
	"github.com/msjang/pathguard/internal/trayicon"
)

const repoURL = "https://github.com/msjang/pathguard"

var (
	cfg     config.Config
	cfgPath string
	T       func(string) string

	mu         sync.Mutex // guards scanning flag + lastReport
	scanning   bool
	lastReport string

	statusItem *systray.MenuItem
	overGroup  *slotGroup
	warnGroup  *slotGroup
)

func main() {
	c, path, err := config.Load()
	if err != nil {
		log.Printf("config: %v (using defaults)", err)
	}
	cfg, cfgPath = c, path
	T = i18n.Translator(i18n.Resolve(cfg.UI.Language))
	systray.Run(onReady, func() {})
}

func onReady() {
	systray.SetTooltip(T("tooltip"))
	setIcon(trayicon.Idle)

	header := systray.AddMenuItem("Pathguard", "")
	header.Disable()
	statusItem = systray.AddMenuItem(T("status_idle"), "")
	statusItem.Disable()
	systray.AddSeparator()

	maxInline := cfg.Menu.MaxInline
	if maxInline <= 0 {
		maxInline = 10
	}
	overGroup = newSlotGroup(T("over"), maxInline)
	warnGroup = newSlotGroup(T("warnings"), maxInline)
	systray.AddSeparator()

	scanItem := systray.AddMenuItem(T("scan_now"), "")
	settingsItem := systray.AddMenuItem(T("settings"), "")
	aboutItem := systray.AddMenuItem(T("about"), "")
	systray.AddSeparator()
	quitItem := systray.AddMenuItem(T("quit"), "")

	go func() {
		for range scanItem.ClickedCh {
			go runScan()
		}
	}()
	go func() {
		for range settingsItem.ClickedCh {
			openPath(cfgPath)
		}
	}()
	go func() {
		for range aboutItem.ClickedCh {
			openPath(repoURL)
		}
	}()
	go func() {
		<-quitItem.ClickedCh
		systray.Quit()
	}()

	go runScan()      // scan once at startup
	go scheduleLoop() // then on the configured interval
}

// slotGroup manages a parent menu item plus a fixed pool of reveal-able child
// slots and a "full report" item. On each scan we retint the pool: systray has
// no way to remove items, so we reuse a fixed set and Show/Hide them.
type slotGroup struct {
	parent *systray.MenuItem
	slots  []*systray.MenuItem
	report *systray.MenuItem
	mu     sync.Mutex
	paths  []string
}

func newSlotGroup(label string, maxInline int) *slotGroup {
	p := systray.AddMenuItem(label, "")
	g := &slotGroup{parent: p, paths: make([]string, maxInline)}
	for i := 0; i < maxInline; i++ {
		it := p.AddSubMenuItem("", "")
		it.Hide()
		g.slots = append(g.slots, it)
		idx := i
		go func() {
			for range it.ClickedCh {
				g.mu.Lock()
				path := g.paths[idx]
				g.mu.Unlock()
				if path != "" {
					reveal(path)
				}
			}
		}()
	}
	g.report = p.AddSubMenuItem(T("open_report"), "")
	g.report.Hide()
	go func() {
		for range g.report.ClickedCh {
			mu.Lock()
			rp := lastReport
			mu.Unlock()
			if rp != "" {
				openPath(rp)
			}
		}
	}()
	p.Hide()
	return g
}

func (g *slotGroup) update(label string, rs []scan.Record) {
	n := len(rs)
	if n == 0 {
		g.parent.Hide()
		return
	}
	g.parent.SetTitle(fmt.Sprintf("%s (%d)", label, n))
	g.parent.Show()

	g.mu.Lock()
	k := len(g.slots)
	if n < k {
		k = n
	}
	for i := range g.slots {
		if i < k {
			g.slots[i].SetTitle(fmt.Sprintf("%dB  %s", rs[i].NameNFD, filepath.Base(rs[i].Rel)))
			g.paths[i] = rs[i].Abs
			g.slots[i].Show()
		} else {
			g.paths[i] = ""
			g.slots[i].Hide()
		}
	}
	g.mu.Unlock()

	if n > k {
		g.report.SetTitle(fmt.Sprintf("%s (%d)", T("open_report"), n))
		g.report.Show()
	} else {
		g.report.Hide()
	}
}

func runScan() {
	mu.Lock()
	if scanning {
		mu.Unlock()
		return
	}
	scanning = true
	mu.Unlock()
	defer func() {
		mu.Lock()
		scanning = false
		mu.Unlock()
	}()

	setIcon(trayicon.Scanning)
	statusItem.SetTitle(T("status_scan"))

	if len(cfg.Watch) == 0 {
		statusItem.SetTitle(T("scan_no_folder"))
		setIcon(trayicon.Idle)
		return
	}

	lim := scan.Limits{NameMax: cfg.Limits.NameMax, PathMax: cfg.Limits.PathMax, WarnRatio: cfg.Limits.WarnRatio}
	excl := cfg.ExcludeSet()
	var agg scan.Result
	for _, w := range cfg.Watch {
		res, err := scan.Scan(config.ExpandRoot(w.Root), w.RemotePrefix, lim, excl)
		if err != nil {
			log.Printf("scan %s: %v", w.Root, err)
		}
		agg = merge(agg, res)
	}
	sortWorst(&agg)

	overN := len(agg.NameOver) + len(agg.PathOver)
	warnN := len(agg.NameWarn) + len(agg.PathWarn)

	if rp, err := report.Write(agg, cfg.Limits.NameMax, cfg.Limits.PathMax); err == nil {
		mu.Lock()
		lastReport = rp
		mu.Unlock()
	}

	overGroup.update(T("over"), agg.NameOver)
	warnGroup.update(T("warnings"), agg.NameWarn)

	setIcon(iconFor(overN, warnN))
	statusItem.SetTitle(statusText(overN, warnN, agg.Total))
}

// iconFor maps counts to a color (ADR-0005), precedence red > yellow > green.
func iconFor(over, warn int) trayicon.State {
	th := cfg.Notify.Thresholds
	switch {
	case over >= th.Red:
		return trayicon.Over
	case over >= th.Yellow || (over == 0 && warn >= th.Warn):
		return trayicon.Warn
	default:
		return trayicon.OK
	}
}

func statusText(over, warn, total int) string {
	if over == 0 && warn == 0 {
		return fmt.Sprintf("%s (%d)", T("status_clean"), total)
	}
	return fmt.Sprintf("%s %d · %s %d", T("over"), over, T("warnings"), warn)
}

func scheduleLoop() {
	d, err := time.ParseDuration(cfg.Schedule.Interval)
	if err != nil || d <= 0 {
		d = 6 * time.Hour
	}
	t := time.NewTicker(d)
	defer t.Stop()
	for range t.C {
		runScan()
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

func sortWorst(r *scan.Result) {
	sort.SliceStable(r.NameOver, func(i, j int) bool { return r.NameOver[i].NameNFD > r.NameOver[j].NameNFD })
	sort.SliceStable(r.NameWarn, func(i, j int) bool { return r.NameWarn[i].NameNFD > r.NameWarn[j].NameNFD })
	sort.SliceStable(r.PathOver, func(i, j int) bool { return r.PathOver[i].PathNFD > r.PathOver[j].PathNFD })
	sort.SliceStable(r.PathWarn, func(i, j int) bool { return r.PathWarn[i].PathNFD > r.PathWarn[j].PathNFD })
}

func setIcon(s trayicon.State) { systray.SetIcon(trayicon.PNG(s)) }

// reveal opens the OS file manager with the file selected (ADR-0010).
func reveal(path string) {
	switch runtime.GOOS {
	case "darwin":
		_ = exec.Command("open", "-R", path).Start()
	case "windows":
		_ = exec.Command("explorer", "/select,"+path).Start()
	default: // Linux: no standard file-select; open the containing folder.
		_ = exec.Command("xdg-open", filepath.Dir(path)).Start()
	}
}

// openPath opens a file or URL with the OS default handler.
func openPath(path string) {
	switch runtime.GOOS {
	case "darwin":
		_ = exec.Command("open", path).Start()
	case "windows":
		_ = exec.Command("cmd", "/c", "start", "", path).Start()
	default:
		_ = exec.Command("xdg-open", path).Start()
	}
}
