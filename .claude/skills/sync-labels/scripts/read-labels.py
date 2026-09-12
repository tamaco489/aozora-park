#!/usr/bin/env python3
"""labels.md のラベル定義を JSON で出力する

sync-labels スキルの Step 2 と Step 8 が使う。定義を読むのはこの 1 箇所にする。

使い方:
    .claude/skills/sync-labels/scripts/read-labels.py [定義ファイル]

既定の定義ファイルは .claude/rules/github/labels.md。
出力は [{"name": ..., "color": ..., "description": ...}, ...] で、色は # を外した 6 桁、
説明はバッククォートを外した「対象スコープ」の文言。
"""

import json
import pathlib
import re
import sys

DEFAULT_PATH = ".claude/rules/github/labels.md"

# | `frontend` | `#fb10a9` | フロントエンド (`frontend/` 配下) |
ROW = re.compile(r"^\|\s*`([^`]+)`\s*\|\s*`#([0-9a-fA-F]{6})`\s*\|\s*(.+?)\s*\|$")

# GitHub のラベル説明の上限
MAX_DESCRIPTION = 100


def main():
    path = pathlib.Path(sys.argv[1] if len(sys.argv) > 1 else DEFAULT_PATH)
    if not path.exists():
        print(f"定義ファイルが見つかりません: {path}", file=sys.stderr)
        sys.exit(1)

    labels = []
    for line in path.read_text().splitlines():
        matched = ROW.match(line)
        if not matched:
            continue
        description = matched.group(3).replace("`", "")
        if len(description) > MAX_DESCRIPTION:
            print(f"説明が {MAX_DESCRIPTION} 文字を超えています: {matched.group(1)}", file=sys.stderr)
            sys.exit(1)
        labels.append(
            {"name": matched.group(1), "color": matched.group(2).lower(), "description": description}
        )

    if not labels:
        print(f"ラベルの表を読み取れませんでした: {path}", file=sys.stderr)
        sys.exit(1)

    print(json.dumps(labels, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
