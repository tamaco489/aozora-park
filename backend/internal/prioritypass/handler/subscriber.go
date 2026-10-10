package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/tamaco489/aozora-park/backend/internal/platform/serving/apperr"
	"github.com/tamaco489/aozora-park/backend/internal/platform/serving/pubsubpush"
	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
	prioritypassusecase "github.com/tamaco489/aozora-park/backend/internal/prioritypass/usecase"
)

// requestedMessage は prioritypass.requested から受け取る本文
//
//   - 送る側は internal/prioritypass/infrastructure/pubsub が組み立てる
//   - handler は infrastructure を import しないため同じ項目名を両方に持つ、片方を変更したらもう片方も修正する
type requestedMessage struct {
	PassID string `json:"passId"`
}

// errBadMessage は本文が優先パスの申込として読めないことを示す
//
// 入口で起きる失敗のため domain のセンチネルにはしない
var errBadMessage = apperr.New(
	apperr.KindInvalidArgument,
	"PRIORITY_PASS_BAD_MESSAGE",
	"優先パスの申込として読めないメッセージ",
)

// NewSubscriber は申込の割当を起動する push の受け口を組み立てる
//
// エンベロープと応答の扱いは pubsubpush が持つため、ここは本文の翻訳とユースケースの呼び出しだけを行う
func NewSubscriber(
	logger *slog.Logger,
	allocateTimeSlot *prioritypassusecase.AllocateTimeSlot,
) http.Handler {
	return pubsubpush.NewHandler(
		logger,
		func(ctx context.Context, msg *pubsubpush.Message) error {
			var body requestedMessage
			if err := json.Unmarshal(msg.Data, &body); err != nil {
				// 同じ本文が何度届いても解釈できないため、再配信させずに ack する
				return errBadMessage
			}

			if body.PassID == "" {
				return errBadMessage
			}

			_, err := allocateTimeSlot.Do(
				ctx,
				prioritypassusecase.AllocateTimeSlotInput{
					PassID: prioritypassmodel.PassID(body.PassID),
				},
			)

			return err
		})
}
