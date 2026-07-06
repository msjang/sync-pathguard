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

- **T-0004** `TODO` — Go 프로젝트 스캐폴드: `go.mod`, 디렉터리 구조, 교차컴파일 스크립트(darwin arm64/amd64, windows amd64). (ADR-0002)
- **T-0005** `TODO` — 스캔 코어 이식: `pathguard.py` 로직을 Go로. `x/text/unicode/norm`으로 NFD 바이트 측정, NAME_MAX/PATH_MAX 판정. Python 대비 결과 동일성 검증. (ADR-0001, 0002)
- **T-0006** `TODO` — YAML 설정 로더: 스키마(watch/limits/schedule/notify/exclude/ui) 파싱, 기본값, OS별 설정 경로. 없으면 기본 config 생성. (ADR-0003, 0008, 0009)
- **T-0006b** `TODO` — i18n message catalog: `ui.language = auto|en|ko`, auto는 시스템 로케일. 메뉴/알림 문자열 카탈로그화. (ADR-0009)

## 상주 앱

- **T-0007** `TODO` — systray 상주 뼈대: 맥 메뉴바 + 윈도우 트레이 아이콘 표시, 기본 메뉴 골격. (ADR-0004)
- **T-0007b** `TODO` — 메뉴 구조 구현: 상태요약 헤더 + 초과/경고 서브메뉴 트리 + Scan/Settings/About/Quit. (ADR-0010)
- **T-0007c** `TODO` — 결과 탐색(reveal): 파일 항목 클릭 시 macOS `open -R` / Windows `explorer /select` / Linux 폴백. (ADR-0010, OBS-...-05)
- **T-0007d** `TODO` — `Settings…`: conf.yml을 OS 기본 에디터로 열기(전용 UI는 후속). (ADR-0010)
- **T-0007e** `TODO` — `지금 검사(Scan now)` + 검사 중 blue 아이콘. (ADR-0005, 0010)
- **T-0007f** `TODO` — 결과 리포트 파일 생성(HTML/txt): 초과 `>= menu.max_inline`일 때 "전체 리포트 열기…". (ADR-0010)
- **T-0008** `TODO` — 스케줄러: `schedule.interval`(+선택 `at`) 기준 폴링 스캔 루프. (ADR-0004)
- **T-0009** `TODO` — 감시 폴더 관리 UI: 트레이 메뉴에서 폴더 추가/선택 + 제외 목록 편집, 설정 파일에 왕복 반영. (ADR-0003, 0004, 0008)
- **T-0009b** `TODO` — 언어 설정 UI: 트레이 메뉴/설정에서 auto/en/ko 선택. (ADR-0009)
- **T-0010** `TODO` — 감시 주기·시각 설정 UI: 메뉴에서 interval/at 조정. (ADR-0004)

## 알림

- **T-0011** `TODO` — 아이콘 색 5-state(gray/blue/green/yellow/red) 렌더 및 전환 로직.
  심볼 = FontAwesome ruler-horizontal. idle=gray, 검사중=blue. 결과색 우선순위 red>yellow>green:
  red=초과≥`thresholds.red`, yellow=초과≥`thresholds.yellow` 또는 경고≥`thresholds.warn`, green=그 외.
  FontAwesome CC BY 4.0 attribution 처리. (ADR-0005, OBS-...-01)
- **T-0012** `TODO` — (선택) OS 네이티브 배너 병행: 맥 UserNotifications, 윈도우 toast. `notify.native_banner` 게이트. (ADR-0005)
- **T-0013** `TODO` — 마지막 스캔 결과(초과/경고 건수) 상태 요약 헤더 노출. (ADR-0010, T-0007b와 연계)

## 배포

- **T-0014** `TODO` — 릴리스 파이프라인: 3타깃 바이너리 빌드·아티팩트 패키징(맥 .app/dmg, 윈도우 .exe).
  표시명 메타데이터 = `Sync Pathguard`, 식별자 = `sync-pathguard`. (ADR-0011)
- **T-0015** `BLOCKED` — 맥 코드서명·공증, 윈도우 코드서명, 로그인 자동시작 등록. (ADR-0007, 인증서/계정 필요)

## 정리

- **T-0016** `TODO` — Python `pathguard.py`의 위치 정리: Go 이식 검증 후 참조용 CLI로 남길지/제거할지 결정.
