// Package main は定期ジョブをサブコマンドで切り替える 1 つのバイナリ
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tamaco489/aozora-park/backend/internal/platform/client/firestore"
	"github.com/tamaco489/aozora-park/backend/internal/platform/config"
	"github.com/tamaco489/aozora-park/backend/internal/platform/observability/logging"
)

// generateSubcommand は枠を作成するサブコマンド、Cloud Run jobs からは引数で渡す
const generateSubcommand = "generate"

// jobTimeout はジョブ全体の締め切り
//
//   - Firestore のクライアントは到達できない相手に対して再試行を続けるため、期限が無いと失敗せずに走り続ける
//   - 障害時に Cloud Run jobs のタスクのタイムアウトまで失敗が検知されず、再試行の開始もその分だけ遅れる
//   - 枠の生成は 150 件の dateInventories と 780 件の timeSlots を 2 秒で書き終えたため、桁違いの余裕を見てこの値にする
const jobTimeout = 5 * time.Minute

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatalf("job: %v", err)
	}
}

func run(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: job %s", generateSubcommand)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	logger := logging.New(cfg.LogLevel)

	ctx, cancel := context.WithTimeout(context.Background(), jobTimeout)
	defer cancel()

	firestoreClient, err := firestore.New(ctx, cfg.ProjectID)
	if err != nil {
		return err
	}
	defer func() {
		if err := firestoreClient.Close(); err != nil {
			logger.Error("close firestore",
				"error",
				err,
			)
		}
	}()

	switch args[0] {
	case generateSubcommand:
		return generate(
			ctx,
			logger,
			firestoreClient,
		)
	default:
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
}
