#!/usr/bin/env python3
"""GitHub に書き込む操作とコミットに Claude のセッション URL が含まれていないかを検査する PreToolUse フック

対象は Bash の git commit / gh pr / gh issue / gh api と、GitHub に書き込む MCP のツール。
コマンド本体 (ヒアドキュメントを含む) に加え、--body-file / -F / --file= / -F body=@ で渡したファイルの中身も見る。
コマンド置換・リダイレクト・パイプで渡した本文や、cd した先の相対パスは追わない。
含まれていれば permissionDecision: deny を返してやり直させる。
"""

import json
import re
import shlex
import sys
from pathlib import Path

PATTERN = "claude.ai/code/session"

# git の前置オプションは値を取るもの (-C dir など) を 2 語として読み飛ばす
TARGET = re.compile(
    r"(?:^|[;&|(]|\n)\s*"
    r"(?:git\s+(?:(?:-C|-c|--git-dir|--work-tree|--namespace)\s+\S+\s+|-\S+\s+)*commit\b"
    r"|gh\s+(?:pr|issue|api)\b)"
)

# 値がそのままファイルのパスになるフラグ
PATH_FLAGS = ("--body-file", "--input", "--file")

# gh pr / gh issue / git commit では -F がファイル、gh api の -F / --field は key=value か key=@path
FIELD_FLAGS = ("-F", "--field")

# GitHub MCP 以外で GitHub に書き込む MCP のツール
OTHER_GITHUB_TOOLS = (
    "mcp__claude_ai_Mermaid_Chart__create_pr",
    "mcp__claude_ai_Mermaid_Chart__create_issue",
    "mcp__claude_ai_Mermaid_Chart__push_file",
)


def allow():
    print(json.dumps({}))
    sys.exit(0)


def deny():
    print(
        json.dumps(
            {
                "hookSpecificOutput": {
                    "hookEventName": "PreToolUse",
                    "permissionDecision": "deny",
                    "permissionDecisionReason": (
                        "Claude のセッション URL (claude.ai/code/session_...) が含まれています。\n"
                        "コミットメッセージにも Issue・PR の本文とコメントにも書かないでください。\n"
                        "PR 本文の末尾は 🤖 Generated with [Claude Code](https://claude.com/claude-code) の 1 行だけにします。"
                    ),
                }
            },
            ensure_ascii=False,
        )
    )
    sys.exit(0)


def referenced_files(command: str) -> list[str]:
    """コマンドに渡されたファイルのパスを取り出す"""
    try:
        tokens = shlex.split(command)
    except ValueError:
        return []

    paths = []
    for i, token in enumerate(tokens):
        flag, sep, inline = token.partition("=")
        if flag in PATH_FLAGS and sep:
            paths.append(inline)
            continue

        if i + 1 >= len(tokens):
            continue
        value = tokens[i + 1]
        if token in PATH_FLAGS:
            paths.append(value)
        elif token in FIELD_FLAGS:
            if "=@" in value:
                paths.append(value.split("=@", 1)[1])
            elif "=" not in value:
                paths.append(value)
    return [path for path in paths if path != "-"]


def main():
    try:
        payload = json.load(sys.stdin)
    except json.JSONDecodeError:
        allow()

    tool = payload.get("tool_name", "")
    tool_input = payload.get("tool_input", {})

    if tool.startswith("mcp__github__") or tool in OTHER_GITHUB_TOOLS:
        if PATTERN in json.dumps(tool_input, ensure_ascii=False):
            deny()
        allow()

    command = tool_input.get("command", "")
    if not TARGET.search(command):
        allow()

    if PATTERN in command:
        deny()

    for path in referenced_files(command):
        try:
            if PATTERN in Path(path).expanduser().read_text(encoding="utf-8"):
                deny()
        except (OSError, UnicodeDecodeError):
            continue

    allow()


if __name__ == "__main__":
    main()
