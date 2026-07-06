#!/usr/bin/env python3
"""스모크 테스트: 임시 폴더에 NFD 초과 파일을 만들고 pathguard가 잡는지 확인.
외부 의존성 없이 표준 라이브러리만 사용 (CI에서 pytest 없이 실행)."""
import json, os, subprocess, sys, tempfile, unicodedata, pathlib

ROOT_DIR = pathlib.Path(__file__).resolve().parent.parent
SCRIPT = ROOT_DIR / "pathguard.py"


def run(root, **env):
    e = dict(os.environ)
    e["PATHGUARD_ROOT"] = str(root)
    e.update({k: str(v) for k, v in env.items()})
    out = subprocess.check_output([sys.executable, str(SCRIPT), "--json"], env=e, text=True)
    return json.loads(out.strip().splitlines()[-1])


def main():
    with tempfile.TemporaryDirectory() as d:
        root = pathlib.Path(d)
        # 정상(짧은) 파일
        (root / unicodedata.normalize("NFC", "정상.txt")).write_text("ok", encoding="utf-8")
        # 받침 글자 반복 → NFD 9B/자. 30자 = 270B > NAME_MAX(255) → 초과여야 함
        long_name = unicodedata.normalize("NFC", ("강" * 30) + ".txt")
        (root / long_name).write_text("x", encoding="utf-8")

        s = run(root, PATHGUARD_NAME_MAX=255, PATHGUARD_PATH_MAX=4096,
                PATHGUARD_WARN=0.80, PATHGUARD_REMOTE_PREFIX="/remote/prefix")

        assert s["total"] >= 2, f"total 예상 >=2, got {s}"
        assert s["name_over"] >= 1, f"NFD 초과 미탐지: {s}"
        print("smoke ok:", s)


if __name__ == "__main__":
    main()
