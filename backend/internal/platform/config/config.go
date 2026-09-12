// Package config は環境変数をまとめて読む
package config

import (
	"fmt"
	"log/slog"
	"os"
)

// Cloud Run は待ち受けポートを PORT で渡す
const defaultPort = "8080"

// Config はプロセス全体で使う設定
type Config struct {
	Port     string     // Port は HTTP サーバの待ち受けポート
	LogLevel slog.Level // LogLevel は構造化ログを出力する下限
}

// Load は環境変数を読んで Config を組み立てる
func Load() (*Config, error) {
	cfg := &Config{
		Port:     os.Getenv("PORT"),
		LogLevel: slog.LevelInfo,
	}

	if cfg.Port == "" {
		cfg.Port = defaultPort
	}

	if v := os.Getenv("LOG_LEVEL"); v != "" {
		if err := cfg.LogLevel.UnmarshalText([]byte(v)); err != nil {
			return nil, fmt.Errorf("LOG_LEVEL %q: %w", v, err)
		}
	}

	return cfg, nil
}
