// Package main は定期ジョブをサブコマンドで切り替える 1 つのバイナリ
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/tamaco489/aozora-park/backend/internal/platform/client/firestore"
	"github.com/tamaco489/aozora-park/backend/internal/platform/config"
	"github.com/tamaco489/aozora-park/backend/internal/platform/observability/logging"
)

// generateSubcommand は枠を作成するサブコマンド、Cloud Run jobs からは引数で渡す
const generateSubcommand = "generate"

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

	ctx := context.Background()

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
