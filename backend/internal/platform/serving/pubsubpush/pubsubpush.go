// Package pubsubpush は Pub/Sub push と機能側のハンドラの間でエンベロープと応答を変換する
//
// Pub/Sub の決まりごと (POST 固定、data が base64、ack は 2xx、再配信は 5xx) をこの層に閉じる
// 機能側が書くのは func(ctx, *Message) error だけで、HTTP のステータスコードを組み立てない
package pubsubpush

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/tamaco489/aozora-park/backend/internal/platform/serving/apperr"
)

// maxBodyBytes は読み込むエンベロープの上限 (16 MiB)
//
// push のリクエストサイズの上限は公式ドキュメントに記載が無いため、以下の通りエンベロープに入りうる最大から見積もる
//   - data は 10 MB まで
//   - 属性は 100 件まで、キーは 256 バイト、値は 1024 バイトまで
//   - base64 で約 4/3 に膨らむ
//
// NOTE: https://docs.cloud.google.com/pubsub/quotas
const maxBodyBytes = 16 << 20

// Message はエンベロープから取り出した 1 件のメッセージ
//
// 再配信は別のリクエストとして届き、前回の処理の記憶はプロセスに残らない
// 何回目かを知る手がかりは DeliveryAttempt だけになる
//
// 不変条件を持たない入れ物のためフィールドを公開する
type Message struct {
	ID              string            // ID は Pub/Sub が採番した messageId
	Data            []byte            // Data は本文、base64 から戻したもの
	Attributes      map[string]string // Attributes は publish したときに付けた属性
	PublishTime     time.Time         // PublishTime は Pub/Sub が受け付けた時刻
	DeliveryAttempt int               // DeliveryAttempt は何回目の配信か、DLQ を設定した購読でだけ入る
	Subscription    string            // Subscription は配信元の購読の完全名
}

// Handler は 1 件のメッセージを処理する
//
// 失敗は apperr のセンチネルで返す、ack と再配信のどちらになるかは apperr.Retryable が決める
// 成否の判断に使わない情報 (件数など) は、ここでログに残してから nil を返す
type Handler func(ctx context.Context, msg *Message) error

// envelope は push が送ってくる JSON の形
type envelope struct {
	Message         envelopeMessage `json:"message"`
	Subscription    string          `json:"subscription"`
	DeliveryAttempt int             `json:"deliveryAttempt"` // DLQ を設定した購読でだけ入る
}

// envelopeMessage はエンベロープが包む 1 件のメッセージ
//
// Data が []byte なのは encoding/json が base64 を自動で戻すため
type envelopeMessage struct {
	Data        []byte            `json:"data"`
	MessageID   string            `json:"messageId"`
	Attributes  map[string]string `json:"attributes"`
	PublishTime time.Time         `json:"publishTime"`
}

// NewHandler は push を受ける http.Handler を組み立てる
//
// 応答は Handler が返したエラーだけで決まる
//   - nil: ack する
//   - 再実行で直りうる失敗: 5xx を返して再配信させ、max_delivery_attempts を超えたら DLQ へ送らせる
//   - それ以外: ack してログに残す (再送しても直らないものを DLQ まで引きずらない)
//
// 再配信そのものは Pub/Sub が行い、1 回の配信が 1 回のリクエストになる
// 間隔と回数はこのプロセスに無く、infra/modules/pubsub/main.tf の retry_policy と dead_letter_policy が決める
//
// 呼び出し元が Pub/Sub であることは Cloud Run の IAM と OIDC トークンが保証するため、ここでは検証しない
func NewHandler(logger *slog.Logger, h Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			// push は POST で固定されるため、他のメソッドは Pub/Sub 以外からの誤った呼び出しになる
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)

			return
		}

		// Pub/Sub の仕様上ありえない大きさは読む前に切る、w を渡すと上限超過でコネクションも閉じる
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

		msg, err := decode(r)
		if err != nil {
			// 同じものが何度送られても解釈できないため、再配信させずに ack する
			logger.ErrorContext(
				r.Context(),
				"Pub/Sub のエンベロープを解釈できない",
				slog.Any("error", err),
			)
			w.WriteHeader(http.StatusNoContent)

			return
		}

		if err := h(r.Context(), msg); err != nil {
			attrs := []slog.Attr{
				slog.String("messageId", msg.ID),
				slog.String("subscription", msg.Subscription),
				slog.Int("deliveryAttempt", msg.DeliveryAttempt),
				slog.Any("error", err),
			}

			// 直らないものを 5xx で返すと、同じ失敗を max_delivery_attempts 回繰り返してから DLQ に入る
			// 結果が変わらないので試行が無駄になるうえ、再投入すべきものと捨てるものが DLQ に混ざる
			if !apperr.Retryable(err) {
				logger.LogAttrs(
					r.Context(),
					slog.LevelError,
					"Pub/Sub のメッセージを破棄、再実行しても直らないため再配信させない",
					attrs...,
				)
				w.WriteHeader(http.StatusNoContent)

				return
			}

			logger.LogAttrs(
				r.Context(),
				slog.LevelWarn,
				"Pub/Sub のメッセージの処理に失敗、再配信させる",
				attrs...,
			)
			http.Error(w,
				"retry",
				http.StatusInternalServerError,
			)

			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

// decode はエンベロープを読み取る
//
// messageId が無いものはエンベロープとして成立していないため、本文が空でも受け付けるのと区別する
func decode(r *http.Request) (*Message, error) {
	var env envelope
	if err := json.NewDecoder(r.Body).Decode(&env); err != nil {
		return nil, err
	}

	if env.Message.MessageID == "" {
		return nil, errMissingMessageID
	}

	return &Message{
		ID:              env.Message.MessageID,
		Data:            env.Message.Data,
		Attributes:      env.Message.Attributes,
		PublishTime:     env.Message.PublishTime,
		DeliveryAttempt: env.DeliveryAttempt,
		Subscription:    env.Subscription,
	}, nil
}
