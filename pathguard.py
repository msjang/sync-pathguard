#!/usr/bin/env python3
"""경로/파일명 길이 가드 (읽기전용).
한글 NFD(조합형) 최악치 바이트로 환산해 NAME_MAX/PATH_MAX 초과·임박 파일을 찾는다.
동기화 대상 파일시스템(예: NAS의 btrfs/ext4)의 한계: NAME_MAX 255, PATH_MAX 4096 (바이트).
설정은 환경변수로 덮어쓸 수 있다 (PATHGUARD_* 참고).
"""
import os, sys, unicodedata, json

ROOT          = os.path.expanduser(os.environ.get("PATHGUARD_ROOT", "~/Documents"))
REMOTE_PREFIX = os.environ.get("PATHGUARD_REMOTE_PREFIX", "/volume1/homes/johndoe/MyDocuments")  # 원격(NAS/클라우드) 쪽 절대경로 루트
NAME_MAX      = int(os.environ.get("PATHGUARD_NAME_MAX", "255"))    # 바이트, 경로 구성요소(파일/폴더명) 하나당
PATH_MAX      = int(os.environ.get("PATHGUARD_PATH_MAX", "4096"))   # 바이트, 전체 경로
WARN          = float(os.environ.get("PATHGUARD_WARN", "0.80"))     # 한계의 80%부터 경고

def b_nfc(s): return len(unicodedata.normalize('NFC', s).encode('utf-8'))
def b_nfd(s): return len(unicodedata.normalize('NFD', s).encode('utf-8'))
def form(s):
    if s == unicodedata.normalize('NFC', s): return 'NFC'
    if s == unicodedata.normalize('NFD', s): return 'NFD'
    return 'mixed'

def scan(root=ROOT):
    name_over, name_warn, path_over, path_warn = [], [], [], []
    total = 0
    for dp, dns, fns in os.walk(root):
        # 숨김/시스템 폴더(.git, @eaDir 등)도 동기화 대상이라 포함해야 정확 (제외 안 함)
        for name in list(dns) + list(fns):
            total += 1
            full = os.path.join(dp, name)
            rel  = os.path.relpath(full, root)
            nfd_name = b_nfd(name)                    # 구성요소 최악치
            remote = REMOTE_PREFIX + "/" + rel
            nfd_path = b_nfd(remote)                  # 전체경로 최악치(원격)
            rec = {
                'rel': rel, 'form': form(name),
                'name_cur': len(name.encode('utf-8')), 'name_nfc': b_nfc(name), 'name_nfd': nfd_name,
                'path_nfd': nfd_path,
            }
            if nfd_name > NAME_MAX:      name_over.append(rec)
            elif nfd_name >= NAME_MAX*WARN: name_warn.append(rec)
            if nfd_path > PATH_MAX:      path_over.append(rec)
            elif nfd_path >= PATH_MAX*WARN: path_warn.append(rec)
    return total, name_over, name_warn, path_over, path_warn

def main():
    total, no, nw, po, pw = scan()
    no.sort(key=lambda r:-r['name_nfd']); nw.sort(key=lambda r:-r['name_nfd'])
    print(f"스캔: {total}개 항목 (한계 NAME_MAX={NAME_MAX}B, PATH_MAX={PATH_MAX}B, NFD 최악치 기준)")
    print(f"  파일/폴더명 초과(>{NAME_MAX}B): {len(no)}  | 경고({int(NAME_MAX*WARN)}~{NAME_MAX}B): {len(nw)}")
    print(f"  전체경로 초과(>{PATH_MAX}B): {len(po)}  | 경고: {len(pw)}")
    def show(rec):
        return (f"    NFD {rec['name_nfd']:>3}B (현재 {rec['name_cur']}B/{rec['form']}, NFC {rec['name_nfc']}B)"
                f"  {rec['rel']}")
    if no:
        print(f"\n■ NAME_MAX 초과 {len(no)}건:")
        for r in no[:40]: print(show(r))
    if nw:
        print(f"\n■ NAME_MAX 경고 {len(nw)}건 (상위 15):")
        for r in nw[:15]: print(show(r))
    if po:
        print(f"\n■ PATH_MAX 초과 {len(po)}건:")
        for r in po[:20]: print(f"    NFD {r['path_nfd']}B  {r['rel']}")
    # 요약 JSON (알림/스케줄용)
    summary = {'total': total, 'name_over': len(no), 'name_warn': len(nw),
               'path_over': len(po), 'path_warn': len(pw)}
    if len(sys.argv) > 1 and sys.argv[1] == '--json':
        print(json.dumps(summary, ensure_ascii=False))
    return summary

if __name__ == '__main__':
    main()
