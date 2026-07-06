# Sync Pathguard

[English](README.md) · [한국어](README.ko.md)

> A read-only watcher that catches filenames which will **silently break your cloud/NAS sync** — before they do.

[![CI](https://github.com/msjang/sync-pathguard/actions/workflows/ci.yml/badge.svg)](https://github.com/msjang/sync-pathguard/actions/workflows/ci.yml)
![platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-blue)
![license](https://img.shields.io/badge/license-MIT-green)

sync-pathguard scans a synced folder (Synology Drive · Dropbox · Nextcloud · OneDrive · iCloud Drive,
etc.) and flags files whose **name or path byte length** may exceed the destination's `NAME_MAX` /
`PATH_MAX` once Unicode is decomposed to **NFD** — a sync failure waiting to happen. It **only reads**;
it never touches your files.

---

## Why it matters — the NFD byte blowup

Filesystem limits `NAME_MAX 255` / `PATH_MAX 4096` count **UTF-8 bytes, not characters**. And the
byte length of the *same* text depends on its Unicode normalization form. Korean is the worst offender:

| Form | `보고서` ("report") | UTF-8 bytes |
|---|---|---|
| **NFC (precomposed)** | 3 code points `보·고·서` | **9 bytes** (3B each) |
| **NFD (decomposed)** | 6 jamo `ㅂㅗ ㄱㅗ ㅅㅓ` | **18 bytes** (6B each, 9B with a final consonant) |

→ **NFD is 2–3× larger than NFC** (3× for syllables with a final consonant). For a name of Korean
syllables that all carry a final consonant:

- NFC 28 chars = 84 B (plenty of room)
- **NFD 28 chars = 252 B (right at the 255 edge)**

### Not just Korean

The "NFD makes it bigger" effect applies to **any script with combining characters** — Korean is
merely the most extreme:

| Char | Language | NFC → NFD | Ratio |
|---|---|---|---|
| `강` | Korean | 3 → 9 B | **3.0×** |
| `が` | Japanese (voiced kana) | 3 → 6 B | 2.0× |
| `ệ` | Vietnamese | 3 → 5 B | 1.67× |
| `й` | Russian | 2 → 4 B | 2.0× |
| `é` | French / German / Spanish… | 2 → 3 B | 1.5× |
| `中` · `A` · `ก` | Han / Latin / Thai | no change | 1.0× |

The engine is language-agnostic — it just measures NFD bytes — so Vietnamese and French filenames are
caught the same way.

### What happens in the sync pipeline

- Say you sync a local folder ↔ a remote (e.g. `remote:/volume1/homes/johndoe/MyDocuments`).
- The remote filesystem (a NAS's btrfs/ext4, a Linux server, …) typically enforces `NAME_MAX 255`,
  `PATH_MAX 4096` (bytes).
- **macOS stores new Korean filenames as NFD**, and even NFC files can decompose to NFD on edit.
- Mobile viewers and sync clients may each normalize differently.

So **even if it looks short on Windows today** (NFC), the moment it decomposes to NFD somewhere in the
pipeline, the byte length spikes past `NAME_MAX` and the sync errors out. That's why this tool always
measures the **NFD (worst-case) byte length**.

### ⚠️ Normalizing won't fix it

If an over-limit file is already NFC, its NFD-equivalent length still exceeds the limit — so converting
to NFC does nothing. **The only real fix is to rename it shorter** (trimming ~7 Korean chars brings NFD
back under 255).

## Features

- 🔎 **NFD worst-case** — judges every name and full path by its NFD-normalized UTF-8 byte length
- 🌐 **Language-agnostic** — Korean, Vietnamese, accented Latin, and any combining-mark script
- 🧩 **Sync-app agnostic** — Synology Drive, Dropbox, whatever; just point it at the remote path prefix
- 🚫 **Exclusions** — noise dirs like `.git` and `@eaDir` (Synology cache) are skipped by default, configurable
- 🛟 **Read-only** — never modifies, moves, or deletes a file
- 📋 **Report + JSON** — human-readable output plus a summary JSON for scripting
- 🧭 **Roadmap** — resident macOS/Windows tray & menu-bar app, YAML config, icon-state alerts (see [Roadmap](#roadmap))

## Install & use

### Tray / menu-bar app

Prebuilt binaries are on the [Releases](../../releases) page. They are currently
**unsigned**, so macOS and Windows warn on first launch — steps below. On first
run the app writes a default config (**Settings…** in the menu opens it); point
`watch` at your synced folder, then use **Scan now**. The icon color shows the
result and clicking an over-limit file reveals it in Finder/Explorer. It's a
menu-bar agent — look for the ruler icon, there's no Dock icon.

**macOS — Homebrew** (once the [tap](packaging/homebrew) is set up):

```bash
brew tap msjang/tap
brew install --cask sync-pathguard   # clears the quarantine flag for you
```

**macOS — manual:** because the app is unsigned, macOS blocks it with "can't be
opened" until you remove the quarantine flag.

1. From [Releases](../../releases), download `Sync-Pathguard-macos-arm64.zip`
   (Apple Silicon) or `-amd64.zip` (Intel) and unzip it.
2. Move **Sync Pathguard.app** to `/Applications`.
3. Clear quarantine and open:

   ```bash
   xattr -dr com.apple.quarantine "/Applications/Sync Pathguard.app"
   open "/Applications/Sync Pathguard.app"
   ```

   Or without Terminal: right-click the app → **Open** → **Open**; on recent
   macOS go to System Settings → Privacy & Security → **Open Anyway**.

**Windows — manual:** download `Sync-Pathguard-windows-amd64.zip`, unzip, run
`Sync Pathguard.exe`. If SmartScreen warns: **More info → Run anyway**.

**Build from source** (any OS, requires Go 1.25+):

```bash
git clone https://github.com/msjang/sync-pathguard.git
cd sync-pathguard
go run ./cmd/sync-pathguard      # or: scripts/build.sh → bin/sync-pathguard
```

Config file location:

- macOS: `~/Library/Application Support/sync-pathguard/config.yml`
- Windows: `%APPDATA%\sync-pathguard\config.yml`
- Linux: `~/.config/sync-pathguard/config.yml`

### CLI (Go)

A pure-Go, dependency-free binary — handy for one-off checks, scripts, or CI
(it exits non-zero when anything is over the limit).

```bash
go run ./cmd/pathguard          # scan configured watches, print a report
go run ./cmd/pathguard --json   # summary JSON
go run ./cmd/pathguard --root ~/Docs --remote-prefix /volume1/homes/johndoe/MyDocuments
```

## Configuration

Both the app and the CLI read the same YAML config (full schema in
[`prj/ADR.md`](prj/ADR.md), ADR-0003):

```yaml
watch:
  - root: ~/Documents
    remote_prefix: /volume1/homes/johndoe/MyDocuments  # remote absolute root, for PATH_MAX
limits:  { name_max: 255, path_max: 4096, warn_ratio: 0.80 }  # 0.80 = warn from 80%
exclude: [.git, node_modules, "@eaDir", "#recycle"]          # noise dirs to skip
notify:
  thresholds: { yellow: 1, red: 10, warn: 1 }   # icon color by over/warn counts
menu: { max_inline: 10 }                          # worst-N shown inline in the menu
ui:   { language: auto }                          # auto (system locale) | en | ko
```

Default exclusions: `.git`, `node_modules`, `@eaDir`, `#recycle`, `#snapshot`,
`.DS_Store`, `.Trashes`, `.Spotlight-V100`, `.fseventsd`, `$RECYCLE.BIN`, `System Volume Information`

CLI flags: `--config <path>`, `--root <dir>`, `--remote-prefix <p>`, `--json`.

> The remote path length is the bottleneck — set `remote_prefix` to your actual sync target.

## How it works

- Measures every name and full path as **UTF-8 bytes after NFD normalization**
- Name NFD > `NAME_MAX` → **over**; `WARN`…limit → **warning**
- Full path (remote absolute) NFD > `PATH_MAX` → over/warning
- Excluded folders are not descended into; excluded files are skipped

## Roadmap

Growing into a resident tray / menu-bar app for macOS (Intel & Apple Silicon) and Windows:

- **Single-binary distribution** — Go rewrite, one dependency-free executable (macOS arm64/amd64, Windows amd64)
- **Resident tray / menu bar** — click to configure watched folders, interval, schedule, exclusions
- **YAML config file** — multiple watch folders, limits/warn ratio/exclusions/alert style
- **Icon-state alerts** — icon color reflects state: gray (idle), blue (scanning), green (clean), yellow (warnings or some over), red (many over) — all thresholds configurable
- **Localized UI** — follows the system locale by default; English and Korean selectable

Design docs live in [`prj/`](prj/) (PRD · ADR · TASKS · NOTES).

## License

[MIT](LICENSE)
