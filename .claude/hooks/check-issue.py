#!/usr/bin/env python3
"""Issue が .claude/rules/github/issue.md の規約に沿うかを検査する PreToolUse フック

対象は GitHub MCP の issue_write (method: create) と、Bash の gh issue create。
規約違反なら permissionDecision: deny を返してやり直させる。

検査するのはタイトルの形式と、本文にテンプレートの見出しがあることの 2 点。
本文の中身の良し悪しは検査しない。
"""

import re
import sys

from _github_rules import LABEL, LABELS, TYPE, TYPES, allow, check_body, deny, gh_args, load_payload

RULES = "規約: .claude/rules/github/issue.md, .github/ISSUE_TEMPLATE/"

# メイン Issue: [milestone-<番号>] <タイトル>
MAIN = re.compile(r"^\[(?:milestone-\d+|release)\] \S.*$")

# サブ Issue・単体 Issue: [ラベル] [変更種別] <タイトル>
SCOPED = re.compile(rf"^\[{LABEL}\] \[{TYPE}\] \S.*$")

HEADINGS_MAIN = ["## 目的", "## リリースブランチ", "## 完了条件"]
HEADINGS_SCOPED = ["## 概要", "## 作業ブランチ", "## やること", "## 完了条件"]


def check(title: str, body, has_parent: bool):
    if has_parent:
        if not SCOPED.match(title):
            deny(
                f"サブ Issue のタイトルが規約に合っていません: {title!r}\n"
                f"  形式: [ラベル] [変更種別] <タイトル>\n"
                f"  ラベル: {', '.join(LABELS)}\n"
                f"  変更種別: {', '.join(TYPES)}",
                RULES,
            )
        check_body(body, HEADINGS_SCOPED, RULES)
        return

    if MAIN.match(title):
        check_body(body, HEADINGS_MAIN, RULES)
        return

    if SCOPED.match(title):
        check_body(body, HEADINGS_SCOPED, RULES)
        return

    deny(
        f"Issue のタイトルが規約に合っていません: {title!r}\n"
        f"  メイン: [milestone-<番号>] <タイトル>\n"
        f"  サブ・単体: [ラベル] [変更種別] <タイトル>",
        RULES,
    )


def main():
    tool, args = load_payload()

    if tool == "mcp__github__issue_write":
        if args.get("method") != "create":
            allow()
        check(args.get("title", ""), args.get("body", ""), bool(args.get("parent_issue_number")))

    if tool == "Bash":
        values = gh_args(args.get("command", ""), "issue")
        if values is None:
            allow()
        # --body-file の中身はここでは読めないので本文の検査を飛ばす
        body = None if values.get("body_file") else values.get("body", "")
        check(values.get("title", ""), body, False)

    allow()


if __name__ == "__main__":
    sys.exit(main())
