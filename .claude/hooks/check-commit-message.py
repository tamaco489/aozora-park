#!/usr/bin/env python3
"""git commit のメッセージが .claude/rules/github/ の規約に沿うかを検査する PreToolUse フック

stdin に Claude Code のフック入力 JSON を受け取り、stdout に判定結果の JSON を返す。
規約違反のときは permissionDecision: deny を返してコミットをやり直させる。

検査するのは subject (1 行目) の形式と、本文があることの 2 点。
本文の中身の良し悪しは検査しない。
"""

import json
import re
import shlex
import sys

# .claude/rules/github/labels.md
SCOPES = [
    "frontend",
    "backend",
    "infra",
    "ci",
    "cd",
    "docs",
    "chore",
]

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

SUBJECT_RE = re.compile(
    rf"^#(\d+) ({'|'.join(TYPES)}): (\S.*?) \(({'|'.join(SCOPES)})\)$"
)

# 本文とみなさない行 (トレーラ)
TRAILER_RE = re.compile(r"^[A-Za-z-]+: .+$")

# .claude/rules/github/commit-subject.md
MAX_SUBJECT_LEN = 50

RULES = (
    "規約: .claude/rules/github/commit-subject.md, commit-types.md, labels.md\n"
    "  形式: #<Issue 番号> <type>: <subject> (<スコープ>)\n"
    f"  type: {', '.join(TYPES)}\n"
    f"  スコープ: {', '.join(SCOPES)}\n"
    "  subject: 体言止めまたは「〜する」形、末尾に句点をつけない\n"
    "  本文: 必須。空行をあけて「何が問題だったか → どう変えたか」を書く"
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


HEREDOC = re.compile(r"<<-?\s*'?\"?(\w+)'?\"?\n(.*?)\n\1", re.DOTALL)


def strip_heredocs(command: str) -> str:
    """ヒアドキュメントの中身を除いたシェルの本体を返す

    本文に `git commit` を含む文書を書く場合など、データ部分を命令と誤認しないようにする。
    """
    return HEREDOC.sub("<<HEREDOC", command)


def extract_message(command: str) -> str | None:
    """git commit の -m / --message に渡された文字列を取り出す"""
    # heredoc 形式 (git commit -m "$(cat <<'EOF' ... EOF)")
    heredoc = HEREDOC.search(command)
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
    # コマンドの先頭かシェルの区切りの直後にあるものだけを対象にする (文字列に含むだけの誤検知を避ける)
    if not re.search(
        r"(?:^|[;&|(]|\n)\s*git\s+(?:-\S+\s+|--\S+\s+)*commit\b", strip_heredocs(command)
    ):
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

    matched = SUBJECT_RE.match(subject)
    if not matched:
        deny(f"コミットメッセージの subject が規約に合っていません: {subject!r}")

    if matched.group(3).endswith(("。", ".")):
        deny(f"subject の末尾に句点をつけないでください: {subject!r}")

    lines = message.strip().splitlines()[1:]
    body = [line for line in lines if line.strip() and not TRAILER_RE.match(line.strip())]
    if not body:
        deny(
            "本文がありません。subject の後に空行をあけて、"
            "何が問題だったか・どう変えたかを書いてください。"
        )

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
