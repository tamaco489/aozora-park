package main

import (
	"context"
	"log"
	"net/http"

	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"

	"github.com/tamaco489/aozora-park/backend/internal/platform/config"
	"github.com/tamaco489/aozora-park/backend/internal/platform/observability/logging"
	"github.com/tamaco489/aozora-park/backend/internal/platform/serving/httpx"
	"github.com/tamaco489/aozora-park/backend/internal/platform/serving/interceptor"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("api: %v", err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	logger := logging.New(cfg.LogLevel)

	// TODO: クライアントを作る

	// 共通処理は 1 つにまとめてすべての connect ハンドラに渡す (機能ごとに組み立てない)
	opts := interceptor.All(logger)

	mux := http.NewServeMux()

	// Cloud Run のプローブが叩く口、grpcurl と grpc-health-probe からも同じ形で呼べる
	// 引数のサービス名を増やすと、そのサービス単位でも状態を答えられる
	mux.Handle(grpchealth.NewHandler(grpchealth.NewStaticChecker(), opts))

	// grpcui と buf curl がサービス一覧を引けるようにする
	reflector := grpcreflect.NewStaticReflector(grpchealth.HealthV1ServiceName)

	// v1 と v1alpha の両方を公開するのは、v1alpha しか呼ばないクライアントがまだあるため
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	// サーバを起動
	app := httpx.NewApp(logger)
	return app.Serve(context.Background(), ":"+cfg.Port, mux)
}
