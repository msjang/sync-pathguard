# ADR — sync-pathguard

아키텍처 결정 기록. 상태: `Proposed` / `Accepted` / `Superseded`.

---

## ADR-0001 — 검사 기준은 NFD(조합형) 최악치 바이트
**상태**: Accepted

**맥락**: 한계는 문자 수가 아니라 UTF-8 바이트 수. 한글은 정규화 형태에 따라 바이트가 2~3배 차이.
지금 NFC로 짧아 보여도 파이프라인 어딘가에서 NFD로 풀리면 한계를 넘긴다.

**결정**: 모든 파일/폴더명과 전체 경로를 **NFD로 정규화한 뒤 UTF-8 바이트**로 측정해 판정한다.
현재 형태(NFC/NFD/mixed)는 참고용으로만 표기.

**결과**: 오탐(짧아 보이지만 실제 위험) 없음. 대신 "지금은 NFC라 괜찮은데?"라는 혼란은 문서로 설명.
정규화로는 못 고치므로 해법은 rename 안내뿐임을 명확히 한다.

---

## ADR-0002 — 구현 스택은 Go + systray, 단일 정적 바이너리
**상태**: Accepted

**맥락**: "맥 인텔/애플실리콘 모두 쉽게" + "윈도우는 의존성 없는 단일 exe"가 스택을 사실상 결정한다.
후보: Go / Rust / Python(PyInstaller). Python 번들은 수십 MB에 사실상 인터프리터 동봉이라 목표에 가장 약함.

**결정**: **Go**로 재구현한다.
- 트레이/메뉴바: `fyne.io/systray` (맥 메뉴바 + 윈도우 트레이 공통)
- NFD 정규화: `golang.org/x/text/unicode/norm`
- 설정: `gopkg.in/yaml.v3`
- 배포: `GOOS/GOARCH` 교차컴파일로 `darwin/arm64`, `darwin/amd64`, `windows/amd64` 각각 **단일 바이너리** 산출(런타임 의존성 0).

**결과**: 현행 `pathguard.py`의 스캔 로직을 Go로 이식. Python 버전은 참조/CLI로 당분간 유지.
Rust 대비 트레이 이벤트 루프 구현이 단순. 맥은 필요 시 코드서명/공증은 후속 과제(ADR-0007 참조).

---

## ADR-0003 — 설정 파일은 YAML
**상태**: Accepted

**맥락**: 감시 폴더가 여러 개일 수 있고, 한계·경고비율·감시주기·알림방식 등 설정 항목이 늘어난다.
사람이 직접 열어 고칠 일이 잦다.

**결정**: 설정을 **YAML** 파일로 둔다. 위치는 OS 관례를 따른다
(맥: `~/Library/Application Support/sync-pathguard/config.yml`, 윈도우: `%APPDATA%\sync-pathguard\config.yml`).
트레이 메뉴에서 바꾼 값은 이 파일에 반영(왕복 편집 가능).

**설정 스키마(초안)**:
```yaml
watch:
  - root: ~/Documents                                # 로컬 감시 대상
    remote_prefix: /volume1/homes/johndoe/MyDocuments # 원격(NAS/클라우드) 절대경로(PATH_MAX 계산용)
limits:
  name_max: 255        # 바이트
  path_max: 4096       # 바이트
  warn_ratio: 0.80     # 한계의 80%부터 경고
schedule:
  interval: 6h         # 감시 주기 (예: 30m, 6h)
  at: []               # (선택) 특정 시각 실행 ["09:00", "18:00"]
notify:
  tray_warn_icon: true # 위험 시 아이콘 색 변경
  native_banner: false # OS 네이티브 배너 병행 여부
  thresholds:          # 초과 파일 '건수' 기준 아이콘 색 임계
    yellow: 1          # 초과 이 개수 이상 → yellow
    red: 10            # 초과 이 개수 이상 → red
exclude:               # 스캔 제외 이름 (모든 watch 공통, ADR-0008)
  - .git
  - "@eaDir"           # 시놀로지 캐시(서버 생성)
  - "#recycle"
  - node_modules
ui:
  language: auto        # auto(시스템 로케일) | en | ko  (ADR-0009)
menu:
  max_inline: 10        # 초과 건수 < 이 값이면 메뉴에 파일 트리 표시, 이상이면 리포트 열기 (ADR-0010)
```

**결과**: `warn_ratio`, 다중 `watch` 등 현재 상수를 설정으로 승격. Python 상수 방식은 Superseded 예정.

---

## ADR-0004 — 상주형 트레이/메뉴바 앱 + 폴링 스캔
**상태**: Accepted

**맥락**: 동기화 클라이언트(Synology Drive, Dropbox 등)가 데스크톱에서만 도는 경우가 많아 로컬 상주가 자연스럽다.
실시간 FS 감시(FSEvents/inotify)는 구현 부담이 크고, 이 문제는 초 단위 즉시성이 필요 없다.

**결정**: 백그라운드 상주하며 **설정된 주기/시각에 폴링 스캔**한다.
트레이/메뉴바 아이콘 클릭 시 메뉴로 상태 확인·즉시 스캔·설정·종료 등을 제공한다.
구체적 메뉴 구조와 결과 탐색(reveal)은 **ADR-0010** 참조.

**결과**: 실시간 감시는 범위 밖(ADR로 나중에 재검토 가능). 폴링이라 리소스 부담 낮음.

---

## ADR-0005 — 알림은 트레이 아이콘 색 상태 변화 우선
**상태**: Accepted

**맥락**: 사용자가 원한 방식은 "아이콘 색이 변하는" 저간섭 알림.
아이콘 심볼은 FontAwesome **ruler-horizontal**(solid) — "길이 가드"라는 뜻과 맞음.

**결정**: 단일 심볼(자)에 **색으로 5-state**를 표현한다. 결과 색(green/yellow/red)의 임계는 초과(over) 파일 **건수** 기준.

| 상태 | 색 | 조건 |
|---|---|---|
| idle | gray | 실행 전 / 아직 스캔 안 함 |
| scanning | blue | 검사 중 |
| ok | green | 스캔 완료, 모든 파일이 길이 만족 (초과 0) |
| warn | yellow | 초과 `>= thresholds.yellow` (기본 1) |
| over | red | 초과 `>= thresholds.red` (기본 10) |

검사 중은 gray가 아니라 **blue**로 구분한다(애니메이션은 깔끔히 만들기 어려워 색상 상태로 대체).
임계값은 `conf.yml`(`notify.thresholds`)에서 수정 가능.
`notify.native_banner`가 켜져 있으면 건수 요약 배너를 병행(맥 UserNotifications, 윈도우 toast).

**결과**: 심볼 1종 + 색 5종(gray/blue/green/yellow/red) 렌더 필요.
FontAwesome Free는 CC BY 4.0 → 배포물에 attribution 필요(NOTES 참고).
"경고(80~100%, 아직 초과 아님) 건수"를 색에 반영할지는 OBS-20260707-04로 관리.

---

## ADR-0006 — 읽기전용 원칙 불변
**상태**: Accepted

**맥락**: 도구가 파일을 건드리면 폴더 훼손·동기화 충돌 위험. 해법은 rename이지만 그건 사용자 판단 몫.

**결정**: 앱은 파일을 **읽기만** 한다. 수정/이동/삭제/정규화 변환을 하지 않는다.
위험 파일은 목록·경로로 안내만 하고, rename은 사용자가 직접.

**결과**: 안전. "정규화로 못 고친다"는 설명과 일관.

---

## ADR-0007 — (Proposed) 맥 코드서명·공증, 윈도우 서명/시작프로그램 등록
**상태**: Proposed

**맥락**: 서명 없는 바이너리는 맥 Gatekeeper 경고, 윈도우 SmartScreen 경고가 뜬다.
상주 앱이면 로그인 시 자동 시작 등록도 필요.

**결정(안)**: 배포 단계에서 맥 `codesign`+`notarytool`, 윈도우 코드서명 인증서 적용.
자동 시작은 맥 LaunchAgent, 윈도우 레지스트리 Run 키 또는 시작프로그램.

**결과**: 초기 개발에는 미적용(로컬 실행). 배포 릴리스 시 확정.

---

## ADR-0008 — 노이즈 디렉터리 기본 제외(설정 가능)
**상태**: Accepted

**맥락**: 초기 가정은 "동기화 대상이니 `.obsidian` 등 시스템 폴더도 다 포함"이었으나, 재검토 결과 틀렸다.
- `@eaDir`는 시놀로지가 **서버 쪽**에 만드는 캐시/썸네일이라 로컬 동기화 소스에 있지도 않고 사용자 콘텐츠도 아님.
- `.git`, `node_modules` 등은 보통 동기화에서 제외하거나, 포함돼도 이름이 ASCII/해시라 길이 위험이 아님.
- 이들을 세면 total·경고 수치에 노이즈가 낀다.

**결정**: 기본 제외 목록을 두고 스캔에서 뺀다. 설정(`exclude`, CLI는 `PATHGUARD_EXCLUDE`)으로 대체 가능.
기본값: `.git`, `node_modules`, `@eaDir`, `#recycle`, `#snapshot`, `.DS_Store`, `.Trashes`,
`.Spotlight-V100`, `.fseventsd`, `$RECYCLE.BIN`, `System Volume Information`.
제외 폴더로는 os.walk가 내려가지 않게 prune.

**결과**: 초기 "전부 포함" 가정을 대체. Python 참조 구현에 반영 완료(env 대체 semantics).

---

## ADR-0009 — 문서·UI 다국어 (문서 en/ko 분리, UI는 로케일 기본)
**상태**: Accepted

**맥락**: NFD 폭증은 한글이 극단적일 뿐 결합 문자 언어(베트남어·라틴 악센트 등) 전반의 문제.
검사 엔진은 이미 언어 무관(`b_nfd`는 어떤 문자열이든 처리)이라 i18n이 필요 없다.
남는 건 (1) 문서 언어, (2) 앱 UI 언어.

**결정**:
- 문서: `README.md`(영어, 1차) + `README.ko.md`(한국어). 상단에서 서로 링크. GitHub 공개 리치 우선.
- UI 언어(트레이 앱): 설정 `ui.language = auto | en | ko`. **기본 `auto`(시스템 로케일)**, 영어·한국어 선택.
  문자열이 적으므로(메뉴/알림 십수 개) message catalog로 가볍게 처리. 새 언어는 카탈로그 추가로 확장.
- 엔진: 손대지 않음(이미 언어 무관).

**결과**: README를 en/ko로 분리. "한글 전용"이 아니라 "NFD-safe filename guard"로 포지셔닝.
Go 트레이 앱 단계에서 카탈로그 골격 마련.

---

## ADR-0010 — 트레이 메뉴 구조와 결과 탐색(reveal in Finder/Explorer)
**상태**: Accepted

**맥락**: 트레이 클릭 시 결과를 얼마나·어떻게 보여줄지. 사용자 요청:
상태 요약 + 초과/경고를 서브메뉴 트리로, 파일 클릭 시 파일 매니저로 이동해 해당 파일 선택.

**결정 — 메뉴 구성(위→아래)**:
```
Sync Pathguard              (헤더, 비활성 — 표시 이름, ADR-0011)
⚠ 초과 3 · 경고 8           (상태 요약, 비활성, 색+문구)
──────────
초과 (n)        ▸           (n>0일 때만; n < menu.max_inline이면 파일 트리)
경고 (n)        ▸
──────────
지금 검사  Scan now         (주기 스캔과 별개로 즉시 스캔)
설정…      Settings…        (초기엔 conf.yml을 OS 기본 에디터로 열기)
정보       About
──────────
나가기     Quit
```

**결정 — 결과 트리 & 탐색**:
- 초과/경고 건수 `< menu.max_inline`(기본 10) → 파일 목록을 인라인 서브메뉴로.
  각 항목 = 이름(+NFD 바이트). `>=`이면 거대 네이티브 서브메뉴 대신 **"전체 리포트 열기…"**(생성 HTML/txt).
- 파일 항목 클릭 → 파일 매니저에서 **선택 표시(reveal)**:
  - macOS: `open -R <path>`
  - Windows: `explorer /select,"<path>"`
  - Linux: 파일 선택 표준 없음 → 베스트에포트 폴더 열기(`xdg-open <dir>`), 가능하면 `nautilus/dolphin --select`.
- 검사 중에는 아이콘 blue(ADR-0005).

**결과**: 소량은 즉시 탐색, 다량은 리포트로. Linux reveal 한계는 OBS-20260707-05로 관리.
`Settings…` 전용 UI는 후속 과제(당장은 에디터로 conf 열기).

---

## ADR-0011 — 표시 이름과 식별자 분리
**상태**: Accepted

**맥락**: repo/CLI 이름이 `sync-pathguard`(소문자·하이픈)라 사람이 보는 곳엔 투박하다.

**결정**:
- **표시 이름(사람이 보는 곳) = `Sync Pathguard`**: README 제목, 트레이 메뉴 헤더,
  앱 번들 표시명(macOS `CFBundleName`), 윈도우 실행파일 제품명, About 창, 네이티브 알림 타이틀.
- **식별자(기계가 쓰는 곳) = `sync-pathguard`**: repo 이름, 바이너리/패키지명, 설정 디렉터리
  (`~/Library/Application Support/sync-pathguard`, `%APPDATA%\sync-pathguard`), 번들 ID 요소.

**결과**: 빌드/패키징 단계에서 표시명 메타데이터를 `Sync Pathguard`로 설정(T-0014). 식별자는 그대로.
