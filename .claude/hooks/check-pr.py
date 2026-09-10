#!/usr/bin/env python3
"""PR が .claude/rules/github/pr-description.md の規約に沿うかを検査する PreToolUse フック

対象は GitHub MCP の create_pull_request と、Bash の gh pr create。
規約違反なら permissionDecision: deny を返してやり直させる。

検査するのはブランチ名、タイトルの形式、base の向き、本文の見出しの 4 点。
本文の中身の良し悪しは検査しない。
"""

import re
import sys

from _github_rules import (
    BRANCH,
    LABEL,
    SLUG,
    TYPE,
    allow,
    check_body,
    current_branch,
    deny,
    gh_args,
    load_payload,
)

RULES = "規約: .claude/rules/github/pr-description.md, .github/PULL_REQUEST_TEMPLATE/"

RELEASE = re.compile(rf"^\[Release\] \[{SLUG}\] \S.*$")
SUB = re.compile(rf"^\[sub\] \[{SLUG}\] \S.*$")
SINGLE = re.compile(rf"^\[{LABEL}\] (?:\[issue-\d+\] )?\[{TYPE}\] \S.*$")

HEADINGS_RELEASE = ["## 目的", "## 対応範囲", "## サブイシュー", "## マージ前の確認事項"]
HEADINGS_OTHER = ["## 概要", "## 変更内容", "## 動作確認"]


def check(title: str, body, head: str, base: str):
    if head and not BRANCH.match(head):
        deny(
            f"ブランチ名が規約に合っていません: {head!r}\n"
            f"  形式: <種別>/<main-issue|sub-issue|issue>-<番号>/<内容>\n"
            f"  例: release/main-issue-12/setup, feature/sub-issue-13/buf-init",
            RULES,
        )

    if head.startswith("release/"):
        if not RELEASE.match(title):
            deny(
                f"リリース PR のタイトルが規約に合っていません: {title!r}\n"
                f"  形式: [Release] [識別子] <タイトル>",
                RULES,
            )
        if base and base != "main":
            deny(f"リリース PR の base は main にしてください: {base!r}", RULES)
        check_body(body, HEADINGS_RELEASE, RULES)
        return

    if "sub-issue-" in head:
        if not SUB.match(title):
            deny(
                f"サブ PR のタイトルが規約に合っていません: {title!r}\n"
                f"  形式: [sub] [識別子] <タイトル> (識別子はリリース PR と揃える)",
                RULES,
            )
        if base == "main":
            deny("サブ PR の base はリリースブランチにしてください。main へ直接マージしない。", RULES)
        check_body(body, HEADINGS_OTHER, RULES)
        return

    if not SINGLE.match(title):
        deny(
            f"PR のタイトルが規約に合っていません: {title!r}\n"
            f"  形式: [ラベル] [issue-<番号>] [変更種別] <タイトル> (issue 番号は無ければ省略)",
            RULES,
        )
    check_body(body, HEADINGS_OTHER, RULES)


def main():
    tool, args = load_payload()

    if tool == "mcp__github__create_pull_request":
        check(
            args.get("title", ""),
            args.get("body", ""),
            args.get("head", ""),
            args.get("base", ""),
        )

    if tool == "Bash":
        values = gh_args(args.get("command", ""), "pr")
        if values is None:
            allow()
        # --body-file の中身はここでは読めないので本文の検査を飛ばす
        body = None if values.get("body_file") else values.get("body", "")
        check(
            values.get("title", ""),
            body,
            values.get("head") or current_branch(),
            values.get("base", ""),
        )

    allow()


if __name__ == "__main__":
    sys.exit(main())
