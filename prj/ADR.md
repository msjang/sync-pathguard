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
```

**결과**: `warn_ratio`, 다중 `watch` 등 현재 상수를 설정으로 승격. Python 상수 방식은 Superseded 예정.

---

## ADR-0004 — 상주형 트레이/메뉴바 앱 + 폴링 스캔
**상태**: Accepted

**맥락**: 동기화 클라이언트(Synology Drive, Dropbox 등)가 데스크톱에서만 도는 경우가 많아 로컬 상주가 자연스럽다.
실시간 FS 감시(FSEvents/inotify)는 구현 부담이 크고, 이 문제는 초 단위 즉시성이 필요 없다.

**결정**: 백그라운드 상주하며 **설정된 주기/시각에 폴링 스캔**한다.
트레이/메뉴바 아이콘 클릭 시 메뉴로:
- 감시 폴더 추가/선택,
- 감시 주기·시각 설정,
- 마지막 스캔 결과(초과·경고 건수) 확인,
- 지금 스캔 / 설정 열기 / 종료.

**결과**: 실시간 감시는 범위 밖(ADR로 나중에 재검토 가능). 폴링이라 리소스 부담 낮음.

---

## ADR-0005 — 알림은 트레이 아이콘 색 상태 변화 우선
**상태**: Accepted

**맥락**: 사용자가 원한 방식은 "아이콘 색이 변하는" 저간섭 알림.
아이콘 심볼은 FontAwesome **ruler-horizontal**(solid) — "길이 가드"라는 뜻과 맞음.

**결정**: 단일 심볼(자)에 **색으로 4-state**를 표현한다. 색 임계는 초과(over) 파일 **건수** 기준.

| 상태 | 색 | 조건 |
|---|---|---|
| idle | gray | 실행 전 / 아직 스캔 안 함 |
| ok | green | 스캔 완료, 모든 파일이 길이 만족 (초과 0) |
| warn | yellow | 초과 `>= thresholds.yellow` (기본 1) |
| over | red | 초과 `>= thresholds.red` (기본 10) |

임계값은 `conf.yml`(`notify.thresholds`)에서 수정 가능.
`notify.native_banner`가 켜져 있으면 건수 요약 배너를 병행(맥 UserNotifications, 윈도우 toast).

**결과**: 심볼 1종 + 색 4종(gray/green/yellow/red) 렌더 필요.
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
