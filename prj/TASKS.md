# TASKS — sync-pathguard

상태: `TODO` / `DOING` / `BLOCKED` / `DONE` / `DROP`

## 완료 / 기반

- **T-0001** `DONE` — Python 스캐너(`pathguard.py`): NFD 최악치 바이트로 NAME_MAX/PATH_MAX 검사, 리포트 + `--json`.
- **T-0002** `DONE` — prj 문서 채우기: PRD / ADR(0001~0007) / TASKS / NOTES 초안.
- **T-0003** `DONE` — 개인 흔적 제거 & 일반화: 특정 폴더명/앱 종속 표현 제거,
  `SYNO_PREFIX`→`REMOTE_PREFIX`(config `remote_prefix`), 동기화 앱 무관 서술, 환경변수(`PATHGUARD_*`) 설정화.
- **T-0004a** `DONE` — README를 GitHub 공개 앱 소개형으로 재작성 (뱃지·특징·설치·설정·동작원리·로드맵·라이선스).
- **T-0004b** `DONE` — GitHub Actions CI: Python 문법검사 + 스모크 테스트(임시 폴더에 NFD 초과 파일 생성해 탐지 확인).
- **T-0004c** `DONE` — MIT LICENSE 추가 (holder: sync-pathguard contributors).
- **T-0004d** `DONE` — 노이즈 디렉터리 기본 제외 + `PATHGUARD_EXCLUDE` 설정, Python 반영. (ADR-0008)
- **T-0004e** `DONE` — README를 `README.md`(영어)/`README.ko.md`(한글)로 분리·상호링크, "한글만의 문제 아님" 표 추가. (ADR-0009)

## 스택 전환 (Go)

- **T-0004** `DONE` — Go 스캐폴드: `go.mod`(module `github.com/msjang/sync-pathguard`), `cmd/`·`internal/` 구조, `scripts/build.sh`. (ADR-0002)
- **T-0005** `DONE` — 스캔 코어 이식(`internal/scan`): `x/text/unicode/norm` NFD 바이트, NAME/PATH 판정, worst-first 정렬, 제외. 단위테스트(초과 탐지·제외). (ADR-0001, 0002)
- **T-0006** `DONE` — YAML 설정 로더(`internal/config`): 전체 스키마 파싱, 기본값 병합, OS별 경로, 없으면 생성. (ADR-0003, 0008, 0009)
- **T-0006b** `DONE` — i18n message catalog(`internal/i18n`): `auto|en|ko`, auto=시스템 로케일. 메뉴/상태 문자열. (ADR-0009)

## 상주 앱

- **T-0007** `DONE` — systray 상주 뼈대(`cmd/sync-pathguard`): 트레이/메뉴바 아이콘 + 메뉴 골격. (ADR-0004)
- **T-0007b** `DONE` — 메뉴 구조: 헤더 + 상태요약 + 초과/경고 서브메뉴(worst-first 최대 `menu.max_inline`개
  + 초과분 "전체 리포트 열기…") + Scan/Settings/About/Quit. (ADR-0010)
- **T-0007c** `DONE` — 결과 탐색(reveal): macOS `open -R` / Windows `explorer /select` / Linux 폴더 폴백. (ADR-0010, OBS-...-05)
- **T-0007d** `DONE` — `Settings…`: conf.yml을 OS 기본 앱으로 열기. (ADR-0010)
- **T-0007e** `DONE` — `지금 검사(Scan now)` + 검사 중 blue 아이콘. (ADR-0005, 0010)
- **T-0007f** `DONE` — 결과 리포트(HTML, `internal/report`): 심각도순, 경로, 현재/NFD 바이트, "≈N글자 줄이기" 힌트.
  (후속) URL 스킴 reveal. (ADR-0010)
- **T-0008** `DONE` — 스케줄러: `schedule.interval` 폴링 루프(시작 시 1회 + 주기). `at`(특정 시각)은 TODO. (ADR-0004)
- **T-0009** `TODO` — 감시 폴더 관리 UI: 트레이 메뉴에서 폴더 추가/선택 + 제외 편집, 설정 왕복. (현재는 conf 직접 편집) (ADR-0003, 0004, 0008)
- **T-0009b** `TODO` — 언어 설정 UI: 메뉴에서 auto/en/ko 전환. (현재는 conf `ui.language`) (ADR-0009)
- **T-0010** `TODO` — 감시 주기·시각 설정 UI + `schedule.at`(특정 시각) 지원. (ADR-0004)

## 알림

- **T-0011** `DONE` — 아이콘 색 5-state(`internal/trayicon`) + 전환 로직(우선순위 red>yellow>green,
  경고 반영). ⚠️ 심볼은 **자체 렌더 자(ruler) placeholder** — 정식 FontAwesome ruler-horizontal 에셋·
  CC BY 4.0 attribution·맥 template 처리는 남음(OBS-...-01). (ADR-0005)
- **T-0012** `TODO` — (선택) OS 네이티브 배너 병행: 맥 UserNotifications, 윈도우 toast. `notify.native_banner` 게이트. (ADR-0005)
- **T-0013** `DONE` — 상태 요약 헤더(초과/경고 건수, "모두 정상") 노출. (ADR-0010)

## 배포

- **T-0014** `TODO` — 릴리스 파이프라인: **OS별 네이티브 빌드**(systray=cgo라 교차컴파일 불가) —
  macOS(arm64/amd64) .app/dmg, Windows .exe. 표시명 = `Sync Pathguard`, 식별자 = `sync-pathguard`. (ADR-0002, 0011)
- **T-0015** `BLOCKED` — 맥 코드서명·공증, 윈도우 코드서명, 로그인 자동시작 등록. (ADR-0007, 인증서/계정 필요)

## 정리

- **T-0016** `TODO` — Python `pathguard.py`의 위치 정리: Go 이식 검증 후 참조용 CLI로 남길지/제거할지 결정.
