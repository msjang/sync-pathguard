# NOTES — pathguard

PRD/PROCESS/TASKS/ADR 어디에도 깔끔히 안 들어가는 관찰과 오픈 이슈.

## 오픈 이슈 (결정 필요)

- **OBS-20260707-01 — 아이콘 에셋 [확정]**: **직접 그린** 자 아이콘 사용(`internal/trayicon`,
  외곽선 + 풀높이 눈금, 5색 런타임 렌더). FontAwesome 에셋 사용안은 **기각**(CC BY attribution 부담 회피, ADR-0005).
  외부 에셋/라이선스 없음. 맥 메뉴바는 template 대신 컬러 아이콘으로 색 상태를 그대로 표현.

- **OBS-20260707-02 — "감시 시간" 의미**: 사용자가 말한 "감시 주기랑 시간"에서
  *주기(interval)* 는 명확하나 *시간(at)* 이 "특정 시각 실행"인지 "특정 시간대에만 감시"인지 확인 필요.
  현재 스키마는 interval + 선택적 at: ["09:00"](특정 시각 실행)으로 잡아둠.

- **OBS-20260707-04 — 경고(warn) 건수의 아이콘 반영 [해결]**: 초과 0이어도 경고 `>= thresholds.warn`(기본 1)
  이면 yellow로 변한다(ADR-0005). 즉 green은 "초과 0 + 경고도 임계 미만"일 때만. 6번째 색 없이 yellow 재활용.

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
- **리네임 sync-pathguard → pathguard**: 초기 이름 `sync-pathguard`를 `pathguard`로 통째 변경(브랜드 "Pathguard").
  - 바이너리: CLI=`pathguard`(cmd/pathguard), GUI=`pathguard-gui`(cmd/pathguard-gui), 앱=`Pathguard.app`.
  - module=`github.com/msjang/pathguard`, 번들ID=`io.github.msjang.pathguard`, config 폴더=`pathguard`.
  - ⚠️ config 폴더명이 바뀌어 구버전(`.../sync-pathguard/`) 설정은 새 위치(`.../pathguard/`)로 안 이어짐(초기라 무시).
  - 필요 외부작업: GitHub repo 이름 변경 + 구 v0.1.0 릴리스/태그 삭제 후 재태그(자산명이 `Sync-Pathguard-*`→`Pathguard-*`).
- **Go 라이브러리 확정**: systray=fyne.io/systray, NFD=golang.org/x/text/unicode/norm, YAML=gopkg.in/yaml.v3.
- **Python↔Go 동일성**: 이식 검증용으로만 사용. 동일 입력에서 결과 일치 확인 후 Python 제거(T-0016).
  회귀는 이제 Go 단위테스트(`internal/scan`)가 담당.
- **red 상태 UX(worst-first)**: 초과 10↑이면 "리포트만" 열던 초기안은 구멍 — 브라우저로 연 정적 리포트에선
  파일 매니저 reveal이 안 됨(`file://`은 explorer/select 못 부름). 그래서 메뉴에 **worst N개 인라인 reveal**을
  항상 유지하고 나머지는 리포트로. 사람이 한 번에 다 못 고치므로 worst-first가 실제 작업 순서와도 맞음(ADR-0010).
- **트림 힌트 계산**: 초과 바이트 = `name_nfd - NAME_MAX`. 한글 받침글자 NFD 9B/자이므로 대략 `ceil(초과B/9)`글자
  줄이면 안전(받침 없으면 6B/자). 리포트에 "≈N글자 줄이세요"로 표기(T-0007f).
- **대안(보류)**: 폴더별 중첩 서브메뉴로 전체를 reveal 가능하게 하는 방식도 있으나, 동적 중첩 구현 부담·깊은
  탐색 불편으로 worst-first + 리포트를 택함. 리포트에서 직접 reveal은 URL 스킴으로 후속 확장.

## 배포 주의

- 서명 없는 바이너리 → 맥 Gatekeeper / 윈도우 SmartScreen 경고. 릴리스 전 서명·공증 필요(ADR-0007).
- 상주 앱은 로그인 자동시작 등록 필요: 맥 LaunchAgent, 윈도우 Run 키/시작프로그램.
