# NOTES — sync-pathguard

PRD/PROCESS/TASKS/ADR 어디에도 깔끔히 안 들어가는 관찰과 오픈 이슈.

## 오픈 이슈 (결정 필요)

- **OBS-20260707-01 — 아이콘 에셋**: 심볼은 FontAwesome **ruler-horizontal**(solid)로 결정.
  색 5-state(gray/blue/green/yellow/red)를 렌더해야 함(ADR-0005).
  - 라이선스: FontAwesome Free = CC BY 4.0 → 배포물(앱/문서)에 attribution 표기 필요.
    또는 동일 모양을 직접 그린 SVG로 대체해 라이선스 부담 제거 검토.
  - 맥 메뉴바는 template(단색 자동반전) 관례라 색 표현이 제한적 → 컬러 아이콘 강제 렌더 or 점/뱃지 오버레이 방식 검토.
  - 소스는 SVG 1개(ruler-horizontal) + 색만 바꿔 런타임 렌더 또는 사전 생성 PNG/ICO 5종(gray/blue/green/yellow/red).

- **OBS-20260707-02 — "감시 시간" 의미**: 사용자가 말한 "감시 주기랑 시간"에서
  *주기(interval)* 는 명확하나 *시간(at)* 이 "특정 시각 실행"인지 "특정 시간대에만 감시"인지 확인 필요.
  현재 스키마는 interval + 선택적 at: ["09:00"](특정 시각 실행)으로 잡아둠.

- **OBS-20260707-04 — 경고(warn) 건수의 아이콘 반영**: 색 임계는 초과(over) 건수 기준으로 정함(ADR-0005).
  80~100%(아직 초과는 아님) '경고' 건수도 yellow에 반영할지 미정. 현재는 초과 0이면 green(경고가 있어도).

- **OBS-20260707-05 — Linux reveal 한계**: 파일 매니저에서 '파일 선택'은 macOS(`open -R`)·
  Windows(`explorer /select`)만 표준. Linux는 통일된 방법이 없어 폴더 열기(`xdg-open`)로 폴백,
  일부 매니저만 `--select` 지원. Linux는 상주 앱 1차 범위 밖이라 우선순위 낮음(ADR-0010).

- **OBS-20260707-03 — 동기화 클라이언트는 데스크톱 상주**: 동기화(Synology Drive, Dropbox 등)가
  맥/윈 데스크톱에서만 도는 전제. 모바일/서버 단독 환경은 대상 아님. 감시 대상은 항상 "로컬에 동기화된 폴더".

## 기술 메모

- **NFD 최악치 이유**: macOS가 새 한글 파일명을 NFD로 저장, 수정 시 NFC가 풀리기도 함.
  클라이언트마다 정규화가 달라 "최악치=NFD"로 봐야 안전. 정규화로는 못 고치고 rename만이 해법.
- **바이트 감**: 받침 있는 한글 NFC 28자 = 84B인데 NFD로는 252B(255 코앞). 받침 글자 = NFC 3B → NFD 9B.
- **경로(PATH_MAX) 계산**: 로컬 rel 경로를 remote_prefix + "/" + rel로 이어 원격 절대경로 기준으로 잰다.
  로컬 경로가 아니라 **원격 쪽 경로 길이**가 병목이므로 remote_prefix 정확도가 중요.
- **숨김/시스템 폴더는 기본 제외**(ADR-0008): 초기 "전부 포함" 가정은 틀렸음. `@eaDir`는 시놀 **서버** 캐시라
  로컬 소스에 없고, `.git`/`node_modules`는 보통 미동기화거나 ASCII라 위험 아님. 기본 제외+`exclude` 설정으로 전환.
- **NFD 폭증은 한글 전용 아님**: 베트남어(성조+모음부호 스택)·일본어 탁점가나·러시아어·라틴 악센트도 1.5~2x.
  한글이 3x로 최악일 뿐. 검사 엔진(`b_nfd`)은 언어 무관이라 i18n 불필요, i18n은 문서/UI 언어만 해당(ADR-0009).
- **Go 라이브러리 확정**: systray=fyne.io/systray, NFD=golang.org/x/text/unicode/norm, YAML=gopkg.in/yaml.v3.
- **Python↔Go 동일성**: 이식 시 같은 폴더에 대해 두 구현의 초과/경고 목록이 일치하는지 회귀 확인(T-0005).

## 배포 주의

- 서명 없는 바이너리 → 맥 Gatekeeper / 윈도우 SmartScreen 경고. 릴리스 전 서명·공증 필요(ADR-0007).
- 상주 앱은 로그인 자동시작 등록 필요: 맥 LaunchAgent, 윈도우 Run 키/시작프로그램.
