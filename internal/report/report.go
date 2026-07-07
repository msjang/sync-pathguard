// Package report writes a human-readable HTML report of a scan (T-0007f).
// It includes each offender's path, current/NFD bytes, how many bytes it is
// over, and an approximate "trim N characters" hint.
package report

import (
	"fmt"
	"html"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/msjang/pathguard/internal/scan"
)

// trimChars estimates how many Korean characters to cut to get under the limit.
// A syllable with a final consonant is 9 NFD bytes (worst case), so we divide by 9.
func trimChars(overBytes int) int {
	if overBytes <= 0 {
		return 0
	}
	return int(math.Ceil(float64(overBytes) / 9.0))
}

// Write renders res to an HTML file and returns its path.
func Write(res scan.Result, nameMax, pathMax int) (string, error) {
	var b strings.Builder
	b.WriteString(`<!doctype html><meta charset="utf-8"><title>Pathguard report</title>`)
	b.WriteString(`<style>body{font:14px system-ui,sans-serif;margin:2rem;max-width:60rem}` +
		`h1{font-size:1.3rem}h2{margin-top:2rem}table{border-collapse:collapse;width:100%}` +
		`th,td{border:1px solid #ccc;padding:.3rem .5rem;text-align:left}` +
		`td.n{text-align:right;font-variant-numeric:tabular-nums}` +
		`.over{color:#ea4335;font-weight:600}</style>`)
	b.WriteString("<h1>Pathguard — scan report</h1>")
	fmt.Fprintf(&b, "<p>Total scanned: %d · name over: %d · name warn: %d · path over: %d · path warn: %d</p>",
		res.Total, len(res.NameOver), len(res.NameWarn), len(res.PathOver), len(res.PathWarn))

	nameTable := func(title string, rs []scan.Record) {
		if len(rs) == 0 {
			return
		}
		fmt.Fprintf(&b, "<h2>%s (%d)</h2><table><tr><th>Path</th><th>current</th><th>NFD</th><th>over by</th><th>trim ≈</th></tr>", html.EscapeString(title), len(rs))
		for _, r := range rs {
			over := r.NameNFD - nameMax
			overCell := ""
			if over > 0 {
				overCell = fmt.Sprintf(`<span class="over">+%dB</span>`, over)
			}
			trim := ""
			if t := trimChars(over); t > 0 {
				trim = fmt.Sprintf("%d chars", t)
			}
			fmt.Fprintf(&b, `<tr><td>%s</td><td class="n">%dB</td><td class="n">%dB</td><td class="n">%s</td><td class="n">%s</td></tr>`,
				html.EscapeString(r.Rel), r.NameCur, r.NameNFD, overCell, trim)
		}
		b.WriteString("</table>")
	}
	pathTable := func(title string, rs []scan.Record) {
		if len(rs) == 0 {
			return
		}
		fmt.Fprintf(&b, "<h2>%s (%d)</h2><table><tr><th>Path</th><th>NFD path</th><th>over by</th></tr>", html.EscapeString(title), len(rs))
		for _, r := range rs {
			over := r.PathNFD - pathMax
			overCell := ""
			if over > 0 {
				overCell = fmt.Sprintf(`<span class="over">+%dB</span>`, over)
			}
			fmt.Fprintf(&b, `<tr><td>%s</td><td class="n">%dB</td><td class="n">%s</td></tr>`,
				html.EscapeString(r.Rel), r.PathNFD, overCell)
		}
		b.WriteString("</table>")
	}

	nameTable("NAME_MAX over", res.NameOver)
	nameTable("NAME_MAX warnings", res.NameWarn)
	pathTable("PATH_MAX over", res.PathOver)
	pathTable("PATH_MAX warnings", res.PathWarn)

	path := filepath.Join(os.TempDir(), "pathguard-report.html")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", err
	}
	return path, nil
}
