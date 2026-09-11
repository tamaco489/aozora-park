"""check-issue.py と check-pr.py が共有する規約の定義とヘルパ

このファイル自体はフックではない。`.claude/rules/github/` の規約をコードに写したもの。
"""

import json
import re
import shlex
import subprocess
import sys

LABELS = ["frontend", "backend", "proto", "infra", "ci", "cd", "docs", "claude", "vscode", "repo"]
TYPES = ["feat", "fix", "docs", "refactor", "chore", "add", "remove", "test", "ci"]

LABEL = f"(?:{'|'.join(LABELS)})"
TYPE = f"(?:{'|'.join(TYPES)})"
SLUG = r"[a-z0-9][a-z0-9-]*"

# ブランチ名: <種別>/<main-issue|sub-issue|issue>-<番号>/<内容>
BRANCH = re.compile(rf"^[a-z]+/(?:main-issue|sub-issue|issue)-\d+/{SLUG}$")


def allow():
    print(json.dumps({}))
    sys.exit(0)


def deny(reason: str, rules: str):
    print(
        json.dumps(
            {
                "hookSpecificOutput": {
                    "hookEventName": "PreToolUse",
                    "permissionDecision": "deny",
                    "permissionDecisionReason": f"{reason}\n\n{rules}",
                }
            },
            ensure_ascii=False,
        )
    )
    sys.exit(0)


def check_body(body, headings: list[str], rules: str):
    """本文にテンプレートの見出しが揃っているかを見る"""
    # --body-file 指定などで本文を読めない場合は検査しない
    if body is None:
        return

    if not body:
        deny("本文が空です。テンプレートに従って書いてください。", rules)

    if "<!--" in body:
        deny("本文にテンプレートのコメント (<!-- -->) が残っています。削除してください。", rules)

    missing = [h for h in headings if h not in body]
    if missing:
        deny(
            "本文に必要な見出しがありません: " + ", ".join(missing) + "\n"
            "  テンプレートの見出し構成のまま埋めてください。",
            rules,
        )


def load_payload() -> tuple[str, dict]:
    try:
        payload = json.load(sys.stdin)
    except json.JSONDecodeError:
        allow()
    return payload.get("tool_name", ""), payload.get("tool_input", {})


HEREDOC = re.compile(r"<<-?\s*'?\"?(\w+)'?\"?\n(.*?)\n\1", re.DOTALL)


def gh_args(command: str, subcommand: str) -> dict | None:
    """gh <subcommand> create の引数を取り出す。対象外なら None"""
    # ヒアドキュメントの中身は命令ではないので除く。そのうえでコマンドの位置にあるものだけを対象にする
    bodies = [match.span(2) for match in HEREDOC.finditer(command)]
    pattern = re.compile(rf"(?:^|[;&|(]|\n)\s*gh\s+{subcommand}\s+create\b")
    start = None
    for match in pattern.finditer(command):
        # 一致は直前の区切り文字から始まるため、gh 自体の位置で判定する
        pos = match.start() + match.group(0).index("gh")
        if not any(begin <= pos < end for begin, end in bodies):
            start = pos
            break
    if start is None:
        return None

    # 同じコマンドに複数のヒアドキュメントがある場合に備え、gh より前は読まない
    try:
        tokens = shlex.split(command[start:])
    except ValueError:
        return None

    values = {}
    flags = {"--title": "title", "--body": "body", "--base": "base", "--head": "head"}
    for i, token in enumerate(tokens):
        if token in flags and i + 1 < len(tokens):
            values[flags[token]] = tokens[i + 1]
        elif token == "--body-file":
            values["body_file"] = True
    return values


def current_branch() -> str:
    try:
        out = subprocess.run(
            ["git", "rev-parse", "--abbrev-ref", "HEAD"],
            capture_output=True,
            text=True,
            timeout=5,
        )
    except (OSError, subprocess.SubprocessError):
        return ""
    return out.stdout.strip() if out.returncode == 0 else ""
