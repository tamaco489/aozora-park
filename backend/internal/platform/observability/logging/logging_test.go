package logging

import (
	"log/slog"
	"testing"
)

func TestRename(t *testing.T) {
	tests := map[string]struct {
		groups   []string
		attr     slog.Attr
		wantKey  string
		wantText string
	}{
		"正常系_最上位のlevelの場合_severityに改名されること": {
			attr:     slog.Any(slog.LevelKey, slog.LevelInfo),
			wantKey:  "severity",
			wantText: "INFO",
		},
		"正常系_最上位のmsgの場合_messageに改名されること": {
			attr:     slog.String(slog.MessageKey, "listening"),
			wantKey:  "message",
			wantText: "listening",
		},
		"正常系_警告の場合_値がWARNINGに差し替わること": {
			attr:     slog.Any(slog.LevelKey, slog.LevelWarn),
			wantKey:  "severity",
			wantText: "WARNING",
		},
		"正常系_エラーの場合_値がそのままになること": {
			attr:     slog.Any(slog.LevelKey, slog.LevelError),
			wantKey:  "severity",
			wantText: "ERROR",
		},
		"正常系_入れ子のlevelの場合_改名されないこと": {
			groups:   []string{"request"},
			attr:     slog.Any(slog.LevelKey, slog.LevelInfo),
			wantKey:  slog.LevelKey,
			wantText: "INFO",
		},
		"正常系_対象外のキーの場合_そのままになること": {
			attr:     slog.String("addr", ":8080"),
			wantKey:  "addr",
			wantText: ":8080",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := rename(tt.groups, tt.attr)

			if got.Key != tt.wantKey {
				t.Errorf("rename(%v, %v) のキー = %q, want %q", tt.groups, tt.attr, got.Key, tt.wantKey)
			}
			if got.Value.String() != tt.wantText {
				t.Errorf("rename(%v, %v) の値 = %q, want %q", tt.groups, tt.attr, got.Value.String(), tt.wantText)
			}
		})
	}
}
