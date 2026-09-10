#!/usr/bin/env python3
"""git commit のメッセージが .claude/rules/github/ の規約に沿うかを検査する PreToolUse フック

stdin に Claude Code のフック入力 JSON を受け取り、stdout に判定結果の JSON を返す。
規約違反のときは permissionDecision: deny を返してコミットをやり直させる。

検査対象は subject (メッセージの 1 行目) のみ。
本文・トレーラ (Co-Authored-By など) は検査しない。
"""

import json
import re
import shlex
import sys

# .claude/rules/github/commit-types.md
TYPES = [
    "feat",
    "fix",
    "docs",
    "refactor",
    "chore",
    "add",
    "remove",
    "test",
    "ci",
]

SUBJECT_RE = re.compile(rf"^({'|'.join(TYPES)}): \S.*$")

# .claude/rules/github/commit-subject.md
MAX_SUBJECT_LEN = 50

RULES = (
    "規約: .claude/rules/github/commit-subject.md, commit-types.md\n"
    "  形式: <type>: <日本語の subject>\n"
    f"  type: {', '.join(TYPES)}\n"
    "  subject: 体言止めまたは「〜する」形、末尾に句点をつけない"
)


def allow():
    print(json.dumps({}))
    sys.exit(0)


def deny(reason: str):
    print(
        json.dumps(
            {
                "hookSpecificOutput": {
                    "hookEventName": "PreToolUse",
                    "permissionDecision": "deny",
                    "permissionDecisionReason": f"{reason}\n\n{RULES}",
                }
            },
            ensure_ascii=False,
        )
    )
    sys.exit(0)


def extract_message(command: str) -> str | None:
    """git commit の -m / --message に渡された文字列を取り出す"""
    # heredoc 形式 (git commit -m "$(cat <<'EOF' ... EOF)")
    heredoc = re.search(r"<<-?\s*'?\"?(\w+)'?\"?\n(.*?)\n\1", command, re.DOTALL)
    if heredoc:
        return heredoc.group(2)

    try:
        tokens = shlex.split(command)
    except ValueError:
        return None

    for i, token in enumerate(tokens):
        if token in ("-m", "--message"):
            return tokens[i + 1] if i + 1 < len(tokens) else None
        if token.startswith("--message="):
            return token[len("--message=") :]
        if token.startswith("-m") and len(token) > 2:
            return token[2:]
    return None


def main():
    try:
        payload = json.load(sys.stdin)
    except json.JSONDecodeError:
        allow()

    command = payload.get("tool_input", {}).get("command", "")
    if not re.search(r"\bgit\s+(-\S+\s+|--\S+\s+)*commit\b", command):
        allow()

    # メッセージを変えない・エディタで書く場合は検査しない
    if re.search(r"--no-edit|--amend\s*$|-C\b|--reuse-message|--fixup|--squash", command):
        allow()

    message = extract_message(command)
    if message is None:
        allow()

    subject = message.strip().splitlines()[0].strip() if message.strip() else ""

    if not subject:
        deny("コミットメッセージが空です。")

    if not SUBJECT_RE.match(subject):
        prefix = subject.split(":", 1)[0] if ":" in subject else subject
        deny(
            f"コミットメッセージの subject が規約に合っていません: {subject!r}\n"
            f"  検出した type: {prefix!r}"
        )

    if subject.endswith(("。", ".")):
        deny(f"subject の末尾に句点をつけないでください: {subject!r}")

    if len(subject) > MAX_SUBJECT_LEN:
        print(
            json.dumps(
                {
                    "systemMessage": (
                        f"subject が {len(subject)} 文字です "
                        f"({MAX_SUBJECT_LEN} 文字以内が目安): {subject}"
                    )
                },
                ensure_ascii=False,
            )
        )
        sys.exit(0)

    allow()


if __name__ == "__main__":
    main()
