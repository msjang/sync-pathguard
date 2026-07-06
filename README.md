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

Today it's a single, dependency-free Python 3 script.

```bash
git clone https://github.com/msjang/sync-pathguard.git
cd sync-pathguard

python3 pathguard.py           # human-readable report
python3 pathguard.py --json    # summary JSON (for scheduling/alerts)
```

Point it at your folder and remote path with environment variables:

```bash
PATHGUARD_ROOT="$HOME/Documents" \
PATHGUARD_REMOTE_PREFIX="/volume1/homes/johndoe/MyDocuments" \
  python3 pathguard.py
```

## Configuration

| Variable (env) | Default | Meaning |
|---|---|---|
| `PATHGUARD_ROOT` | `~/Documents` | Local folder to scan |
| `PATHGUARD_REMOTE_PREFIX` | `/volume1/homes/johndoe/MyDocuments` | Remote (NAS/cloud) absolute root, for PATH_MAX |
| `PATHGUARD_NAME_MAX` | `255` | Byte limit for a single file/folder name |
| `PATHGUARD_PATH_MAX` | `4096` | Byte limit for the full path |
| `PATHGUARD_WARN` | `0.80` | Warn once a name reaches 80% of the limit |
| `PATHGUARD_EXCLUDE` | (default list below) | Comma-separated names to skip. **Replaces** the defaults when set; empty string disables exclusion |

Default exclusions: `.git`, `node_modules`, `@eaDir`, `#recycle`, `#snapshot`,
`.DS_Store`, `.Trashes`, `.Spotlight-V100`, `.fseventsd`, `$RECYCLE.BIN`, `System Volume Information`

> The remote path length is the bottleneck — set `PATHGUARD_REMOTE_PREFIX` to your actual sync target.

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
- **Icon-state alerts** — icon color reflects state: gray (idle) → blue (scanning) → green / yellow / red (by over-limit count)
- **Localized UI** — follows the system locale by default; English and Korean selectable

Design docs live in [`prj/`](prj/) (PRD · ADR · TASKS · NOTES).

## License

[MIT](LICENSE)
