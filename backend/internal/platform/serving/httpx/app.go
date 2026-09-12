// Package httpx はサーバの起動と終了処理をまとめる
package httpx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"
)

// shutdownTimeout は SIGTERM 後に処理中のリクエストを待つ上限
const shutdownTimeout = 10 * time.Second

// readHeaderTimeout はヘッダの受信を待つ上限、Slowloris を避けるために必ず設定する
const readHeaderTimeout = 10 * time.Second

type cleanup struct {
	name string
	fn   func(context.Context) error
}

// App はサーバと終了処理をまとめる
type App struct {
	logger   *slog.Logger
	cleanups []cleanup
}

func NewApp(logger *slog.Logger) *App {
	return &App{logger: logger}
}

// Cleanup は終了時に呼ぶ処理を登録する
//
// 登録の逆順に実行する。依存される側を先に作り後に閉じるため
func (a *App) Cleanup(name string, fn func(context.Context) error) {
	a.cleanups = append(a.cleanups, cleanup{name: name, fn: fn})
}

// Serve は SIGTERM か割り込みを受けるまでリクエストを受け付ける
func (a *App) Serve(ctx context.Context, addr string, h http.Handler) error {
	// gRPC は HTTP/2 を要求するが Cloud Run との通信に TLS を張らないため、平文の HTTP/2 (h2c) を許可する
	// リフレクションとヘルスチェックはこれが無いと繋がらない
	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)

	srv := &http.Server{
		Addr:              addr,
		Handler:           h,
		Protocols:         protocols,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	// Cloud Run は終了時に SIGTERM を送る (ctx が閉じることで停止処理へ進む)
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	// ListenAndServe は塞がるため別の goroutine で動かし、起動時の失敗だけをここへ返す
	listenErr := make(chan error, 1)
	go func() {
		a.logger.InfoContext(ctx, "listening", slog.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			listenErr <- err
		}
	}()

	// 起動に失敗したか、シグナルを受けたかのどちらかで先へ進む
	select {
	case err := <-listenErr:
		return fmt.Errorf("listen and serve: %w", err)
	case <-ctx.Done():
	}

	// 受信済みのシグナルで打ち切られないよう、停止処理は独立した期限で動かす
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()

	// サーバを先に閉じてからクライアントを閉じる、処理中のリクエストが依存先を使うため
	a.logger.InfoContext(shutdownCtx, "shutting down")
	shutdownErr := srv.Shutdown(shutdownCtx)
	a.runCleanups(shutdownCtx)

	if shutdownErr != nil {
		return fmt.Errorf("shutdown: %w", shutdownErr)
	}

	return nil
}

// runCleanups は登録の逆順に終了処理を呼ぶ (1 件失敗しても残りを続ける)
func (a *App) runCleanups(ctx context.Context) {
	for _, c := range slices.Backward(a.cleanups) {
		if err := c.fn(ctx); err != nil {
			a.logger.ErrorContext(ctx, "終了処理に失敗",
				slog.String("name", c.name),
				slog.Any("error", err),
			)
		}
	}
}
