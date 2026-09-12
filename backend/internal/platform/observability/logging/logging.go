// Package logging は構造化ログの設定を持つ
package logging

import (
	"log/slog"
	"os"
)

// New は Cloud Logging が解釈できる JSON のロガーを作る
func New(level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:       level,
		ReplaceAttr: rename,
	}))
}

// rename は slog の既定のキーと値を Cloud Logging の書式に合わせる
//
//   - level を severity に、msg を message に改名する
//   - 改名しないと重大度での絞り込みも、一覧での本文表示も効かない
//   - 警告は値も差し替える、Cloud Logging は WARNING と綴るが slog は WARN と書く
//   - 改名するのは最上位の属性だけにする、入れ子に同じ名前があっても触らない
func rename(groups []string, a slog.Attr) slog.Attr {
	if len(groups) > 0 {
		return a
	}

	switch a.Key {
	case slog.LevelKey:
		a.Key = "severity"
		if lv, ok := a.Value.Any().(slog.Level); ok && lv == slog.LevelWarn {
			a.Value = slog.StringValue("WARNING")
		}
	case slog.MessageKey:
		a.Key = "message"
	}
	return a
}
