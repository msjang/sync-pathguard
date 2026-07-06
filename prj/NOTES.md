# NOTES — sync-pathguard

PRD/PROCESS/TASKS/ADR 어디에도 깔끔히 안 들어가는 관찰과 오픈 이슈.

## 오픈 이슈 (결정 필요)

- **OBS-20260707-01 — 아이콘 에셋**: 심볼은 FontAwesome **ruler-horizontal**(solid)로 결정.
  색 4-state(gray/green/yellow/red)를 렌더해야 함(ADR-0005).
  - 라이선스: FontAwesome Free = CC BY 4.0 → 배포물(앱/문서)에 attribution 표기 필요.
    또는 동일 모양을 직접 그린 SVG로 대체해 라이선스 부담 제거 검토.
  - 맥 메뉴바는 template(단색 자동반전) 관례라 색 표현이 제한적 → 컬러 아이콘 강제 렌더 or 점/뱃지 오버레이 방식 검토.
  - 소스는 SVG 1개(ruler-horizontal) + 색만 바꿔 런타임 렌더 또는 사전 생성 PNG/ICO 4종.

- **OBS-20260707-02 — "감시 시간" 의미**: 사용자가 말한 "감시 주기랑 시간"에서
  *주기(interval)* 는 명확하나 *시간(at)* 이 "특정 시각 실행"인지 "특정 시간대에만 감시"인지 확인 필요.
  현재 스키마는 interval + 선택적 at: ["09:00"](특정 시각 실행)으로 잡아둠.

- **OBS-20260707-04 — 경고(warn) 건수의 아이콘 반영**: 색 임계는 초과(over) 건수 기준으로 정함(ADR-0005).
  80~100%(아직 초과는 아님) '경고' 건수도 yellow에 반영할지 미정. 현재는 초과 0이면 green(경고가 있어도).

- **OBS-20260707-03 — 동기화 클라이언트는 데스크톱 상주**: 동기화(Synology Drive, Dropbox 등)가
  맥/윈 데스크톱에서만 도는 전제. 모바일/서버 단독 환경은 대상 아님. 감시 대상은 항상 "로컬에 동기화된 폴더".

## 기술 메모

- **NFD 최악치 이유**: macOS가 새 한글 파일명을 NFD로 저장, 수정 시 NFC가 풀리기도 함.
  클라이언트마다 정규화가 달라 "최악치=NFD"로 봐야 안전. 정규화로는 못 고치고 rename만이 해법.
- **바이트 감**: 받침 있는 한글 NFC 28자 = 84B인데 NFD로는 252B(255 코앞). 받침 글자 = NFC 3B → NFD 9B.
- **경로(PATH_MAX) 계산**: 로컬 rel 경로를 remote_prefix + "/" + rel로 이어 원격 절대경로 기준으로 잰다.
  로컬 경로가 아니라 **원격 쪽 경로 길이**가 병목이므로 remote_prefix 정확도가 중요.
- **숨김/시스템 폴더 포함**: `.git`, `@eaDir` 등도 동기화 대상이라 스캔에서 제외하지 않는다(현행 유지).
- **Go 라이브러리 확정**: systray=fyne.io/systray, NFD=golang.org/x/text/unicode/norm, YAML=gopkg.in/yaml.v3.
- **Python↔Go 동일성**: 이식 시 같은 폴더에 대해 두 구현의 초과/경고 목록이 일치하는지 회귀 확인(T-0005).

## 배포 주의

- 서명 없는 바이너리 → 맥 Gatekeeper / 윈도우 SmartScreen 경고. 릴리스 전 서명·공증 필요(ADR-0007).
- 상주 앱은 로그인 자동시작 등록 필요: 맥 LaunchAgent, 윈도우 Run 키/시작프로그램.
