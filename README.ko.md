# Sync Pathguard

**한국어** · [English](README.md)

> 한글 파일명이 클라우드/NAS 동기화를 **조용히 깨뜨리기 전에** 잡아내는 읽기전용 감시 도구.

[![CI](https://github.com/msjang/sync-pathguard/actions/workflows/ci.yml/badge.svg)](https://github.com/msjang/sync-pathguard/actions/workflows/ci.yml)
![platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-blue)
![license](https://img.shields.io/badge/license-MIT-green)

동기화 폴더(Synology Drive · Dropbox · Nextcloud · OneDrive · iCloud Drive 등)의 파일명·경로
**바이트 길이**를 감시해, 유니코드 조합형(NFD) 확장 때문에 `NAME_MAX`/`PATH_MAX`를 넘겨
**동기화가 실패할 위험이 있는 파일을 미리 찾아냅니다.** 파일은 절대 건드리지 않고 **읽기만** 합니다.

---

## 왜 필요한가 — NFD 바이트 폭증

파일시스템 한계 `NAME_MAX 255` / `PATH_MAX 4096` 은 **문자 수가 아니라 UTF-8 바이트 수**입니다.
한글은 유니코드 정규화 형태에 따라 바이트 수가 달라집니다:

| 형태 | 예 `보고서` | UTF-8 바이트 |
|---|---|---|
| **NFC (완성형)** | 완성 코드포인트 3개 `보·고·서` | **9 바이트** (자당 3B) |
| **NFD (조합형)** | 자모 6개 `ㅂㅗ ㄱㅗ ㅅㅓ` | **18 바이트** (자당 6B, 받침 있으면 9B) |

→ **NFD는 NFC의 2~3배** (받침 있는 글자는 3배). 받침 있는 한글 이름이라면:

- NFC 28글자 = 84B (여유)
- **NFD 28글자 = 252B (255 코앞)**

### 한글만의 문제가 아닙니다

NFD 분해로 바이트가 늘어나는 현상은 **결합 문자를 쓰는 언어 전반**에 나타납니다.
한글이 가장 극단적일 뿐입니다:

| 문자 | 언어 | NFC → NFD | 배율 |
|---|---|---|---|
| `강` | 한국어 | 3 → 9 B | **3.0x** |
| `が` | 일본어(탁점 가나) | 3 → 6 B | 2.0x |
| `ệ` | 베트남어 | 3 → 5 B | 1.67x |
| `й` | 러시아어 | 2 → 4 B | 2.0x |
| `é` | 프랑스·독일·스페인… | 2 → 3 B | 1.5x |
| `中` · `A` · `ก` | 한자·라틴·태국 | 변화 없음 | 1.0x |

검사 엔진은 언어를 가리지 않고 NFD 바이트만 재므로, 베트남어·프랑스어 파일도 그대로 잡습니다.

### 동기화 파이프라인에서 벌어지는 일

- 로컬 폴더 ↔ 원격(예: `remote:/volume1/homes/johndoe/MyDocuments`)로 동기화한다고 합시다.
- 원격 파일시스템(NAS의 btrfs/ext4, Linux 서버 등)은 보통 `NAME_MAX 255`, `PATH_MAX 4096` (바이트).
- **macOS는 새로 만든 한글 파일명을 NFD로 저장**하고, NFC 파일도 일부 수정 시 NFD로 풀리기도 합니다.
- 안드로이드/아이폰 뷰어, 동기화 클라이언트마다 정규화 방식이 다를 수 있습니다.

즉 **지금 NFC라 윈도우에선 짧아 보여도**, 파이프라인 어딘가에서 NFD로 풀리면 길이가 튀어 `NAME_MAX`를
넘겨 동기화 에러가 납니다. 그래서 이 도구는 **항상 NFD(최악치) 바이트로 환산**해서 검사합니다.

### ⚠️ 정규화로는 못 고칩니다

초과 파일이 이미 NFC여도 NFD 환산 길이가 한계를 넘으면 위험합니다.
**NFC로 바꿔도 소용없고, 유일한 해법은 이름을 짧게 rename** 하는 것입니다
(한글 ~7글자쯤 줄이면 NFD < 255).

## 특징

- 🔎 **NFD 최악치 기준** — 모든 파일/폴더명·전체 경로를 NFD로 환산한 UTF-8 바이트로 판정
- 🌐 **언어 무관** — 한글·베트남어·라틴 악센트 등 결합 문자 전반 탐지
- 🧩 **동기화 앱 무관** — Synology Drive든 Dropbox든, 원격 경로 프리픽스만 지정하면 동작
- 🚫 **제외 설정** — `.git`·`@eaDir`(시놀 캐시) 등 노이즈 디렉터리는 기본 제외, 조정 가능
- 🛟 **읽기전용** — 파일을 수정·이동·삭제하지 않음
- 📋 **리포트 + JSON** — 사람이 읽는 출력과 스크립트용 요약 JSON
- 🧭 **로드맵**: 맥/윈도우 트레이·메뉴바 상주 앱, YAML 설정, 아이콘 상태 알림 (아래 [로드맵](#로드맵))

## 설치 & 사용

두 가지 방법이 있습니다 (서명된 사전 빌드 바이너리는 첫 릴리스에서 제공 예정):

### 트레이/메뉴바 앱 (Go)

Go 1.25+ 필요. 첫 실행 시 기본 설정 파일을 생성하고, 메뉴의 **설정…** 으로 열 수 있습니다.
`watch`를 동기화 폴더로 맞춘 뒤 **지금 검사** 하면 아이콘 색으로 결과가 표시되고,
초과 파일을 클릭하면 Finder/탐색기에서 그 파일을 선택해 줍니다.

```bash
git clone https://github.com/msjang/sync-pathguard.git
cd sync-pathguard
go run ./cmd/sync-pathguard      # 또는: scripts/build.sh → bin/sync-pathguard
```

설정 파일 위치:

- macOS: `~/Library/Application Support/sync-pathguard/config.yml`
- Windows: `%APPDATA%\sync-pathguard\config.yml`
- Linux: `~/.config/sync-pathguard/config.yml`

### CLI (Python)

의존성 없는 단일 스크립트 — 일회성 점검·스크립트·CI에 편리합니다.

```bash
python3 pathguard.py           # 사람이 읽는 리포트
python3 pathguard.py --json    # 요약 JSON (스케줄/알림용)
```

검사할 폴더와 원격 경로는 환경변수로 지정합니다:

```bash
PATHGUARD_ROOT="$HOME/Documents" \
PATHGUARD_REMOTE_PREFIX="/volume1/homes/johndoe/MyDocuments" \
  python3 pathguard.py
```

## 설정

트레이 앱은 위 YAML 설정을 읽습니다 (스키마는 [`prj/ADR.md`](prj/ADR.md) ADR-0003).
Python CLI는 환경변수로 설정합니다:

| 변수 (env) | 기본값 | 의미 |
|---|---|---|
| `PATHGUARD_ROOT` | `~/Documents` | 검사 대상 로컬 폴더 |
| `PATHGUARD_REMOTE_PREFIX` | `/volume1/homes/johndoe/MyDocuments` | 원격(NAS/클라우드) 쪽 절대경로 루트 (PATH_MAX 계산용) |
| `PATHGUARD_NAME_MAX` | `255` | 파일/폴더명 하나의 바이트 한계 |
| `PATHGUARD_PATH_MAX` | `4096` | 전체 경로 바이트 한계 |
| `PATHGUARD_WARN` | `0.80` | 한계의 80%부터 경고 |
| `PATHGUARD_EXCLUDE` | (아래 기본 제외 목록) | 쉼표로 구분한 제외 이름. 설정 시 기본값을 **대체**, 빈 문자열이면 제외 없음 |

기본 제외: `.git`, `node_modules`, `@eaDir`, `#recycle`, `#snapshot`,
`.DS_Store`, `.Trashes`, `.Spotlight-V100`, `.fseventsd`, `$RECYCLE.BIN`, `System Volume Information`

> 원격 경로 길이가 병목입니다. `PATHGUARD_REMOTE_PREFIX`를 실제 동기화 대상 경로와 맞추세요.

## 동작 원리

- 모든 파일/폴더의 이름·전체경로를 **NFD로 정규화한 뒤 UTF-8 바이트**로 측정
- 이름 NFD > `NAME_MAX` → **초과**, `WARN`~한계 → **경고**
- 전체경로(원격 절대경로) NFD > `PATH_MAX` → 초과/경고
- 제외 목록의 폴더로는 내려가지 않고, 제외 파일은 건너뜀

## 로드맵

맥(인텔/애플실리콘)·윈도우에 상주하는 트레이·메뉴바 앱으로 발전시킬 계획입니다.

- **단일 바이너리 배포** — Go 재구현, 의존성 없는 단일 실행파일 (macOS arm64/amd64, Windows amd64)
- **트레이/메뉴바 상주** — 클릭 시 메뉴에서 감시 폴더·주기·시각·제외 설정
- **YAML 설정 파일** — 다중 감시 폴더, 한계/경고비율/제외/알림 방식
- **아이콘 상태 알림** — 아이콘 색: gray(대기)·blue(검사 중)·green(깨끗)·yellow(경고 또는 일부 초과)·red(다수 초과), 임계값 설정 가능
- **다국어 UI** — 시스템 로케일 기본, 영어·한국어 선택

설계 문서는 [`prj/`](prj/) 참고 (PRD · ADR · TASKS · NOTES).

## License

[MIT](LICENSE)
