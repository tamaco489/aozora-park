// Package config は環境変数をまとめて読む
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
)

// Cloud Run は待ち受けポートを PORT で渡す
const defaultPort = "8080"

// Config はプロセス全体で使う設定
type Config struct {
	Port      string     // Port は HTTP サーバの待ち受けポート
	LogLevel  slog.Level // LogLevel は構造化ログを出力する下限
	ProjectID string     // ProjectID は Firestore の接続先となる GCP プロジェクト
}

// Load は環境変数を読んで Config を組み立てる
func Load() (*Config, error) {
	cfg := &Config{
		Port:      os.Getenv("PORT"),
		LogLevel:  slog.LevelInfo,
		ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
	}

	if cfg.Port == "" {
		cfg.Port = defaultPort
	}

	if v := os.Getenv("LOG_LEVEL"); v != "" {
		if err := cfg.LogLevel.UnmarshalText([]byte(v)); err != nil {
			return nil, fmt.Errorf("LOG_LEVEL %q: %w", v, err)
		}
	}

	// 既定値を置かない、接続先を誤ったまま動く余地を残さないため
	if cfg.ProjectID == "" {
		return nil, errors.New("GOOGLE_CLOUD_PROJECT is required")
	}

	return cfg, nil
}
