// Package i18n is a tiny message catalog (ADR-0009). Language is auto (system
// locale), en, or ko. Kept intentionally small — a handful of menu/status strings.
package i18n

import (
	"os"
	"strings"
)

type Lang string

const (
	EN Lang = "en"
	KO Lang = "ko"
)

var catalog = map[string]map[Lang]string{
	"scan_now":       {EN: "Scan now", KO: "지금 검사"},
	"settings":       {EN: "Settings…", KO: "설정…"},
	"about":          {EN: "About", KO: "정보"},
	"quit":           {EN: "Quit", KO: "나가기"},
	"over":           {EN: "Over limit", KO: "초과"},
	"warnings":       {EN: "Warnings", KO: "경고"},
	"open_report":    {EN: "Open full report…", KO: "전체 리포트 열기…"},
	"status_idle":    {EN: "Not scanned yet", KO: "아직 검사 안 함"},
	"status_scan":    {EN: "Scanning…", KO: "검사 중…"},
	"status_clean":   {EN: "All within limits", KO: "모두 정상"},
	"tooltip":        {EN: "Pathguard", KO: "Pathguard"},
	"scan_no_folder": {EN: "No watch folder configured", KO: "감시 폴더가 설정되지 않음"},
}

// Resolve turns a config language value ("auto"|"en"|"ko") into a concrete Lang,
// falling back to the system locale for "auto" and to EN otherwise.
func Resolve(cfgLang string) Lang {
	switch strings.ToLower(cfgLang) {
	case "ko":
		return KO
	case "en":
		return EN
	default: // auto
		if isKoreanLocale() {
			return KO
		}
		return EN
	}
}

func isKoreanLocale() bool {
	for _, k := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := os.Getenv(k); v != "" {
			return strings.HasPrefix(strings.ToLower(v), "ko")
		}
	}
	return false
}

// Translator returns a T(key) function bound to a language.
func Translator(lang Lang) func(string) string {
	return func(key string) string {
		if m, ok := catalog[key]; ok {
			if s, ok := m[lang]; ok {
				return s
			}
			if s, ok := m[EN]; ok {
				return s
			}
		}
		return key
	}
}
