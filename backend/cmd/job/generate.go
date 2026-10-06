package main

import (
	"context"
	"log/slog"

	gcpfirestore "cloud.google.com/go/firestore"

	"github.com/tamaco489/aozora-park/backend/internal/inventory"
)

// generate は今日から inventoryDays 日分の枠を、まだ無いものだけ作成する
func generate(
	ctx context.Context,
	logger *slog.Logger,
	client *gcpfirestore.Client,
) error {
	result, err := inventory.NewGenerateInventory(client).Do(ctx)
	if err != nil {
		return err
	}

	// 既にある枠は数えないため、2 回目以降の実行では 0 件になる
	logger.Info("generated inventory",
		"dateInventories",
		result.DateInventories,
		"timeSlots",
		result.TimeSlots,
	)

	return nil
}
