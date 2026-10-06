// Package pubsub は優先パスの出来事を Pub/Sub へ送り出す
package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	gcppubsub "cloud.google.com/go/pubsub/v2"

	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
)

// requestedTopicID は申込を流すトピック
//
// infra/modules/pubsub/main.tf の google_pubsub_topic.prioritypass_requested と同じ値にする
const requestedTopicID = "prioritypass.requested"

// attributePassID は購読側が本文を読まずに対象を特定できるようにする属性
const attributePassID = "passId"

// requestedMessage は prioritypass.requested に流す本文
//
// 自分たちで定義した形のため、SDK との間に翻訳の層を挟まず JSON にして送る
type requestedMessage struct {
	PassID       string    `json:"passId"`
	ParkID       string    `json:"parkId"`
	TicketID     string    `json:"ticketId"`
	AttractionID string    `json:"attractionId"`
	TimeSlotID   string    `json:"timeSlotId"`
	RequestedAt  time.Time `json:"requestedAt"`
}

// Publisher は申込を prioritypass.requested へ送る
//
// usecase/port.Publisher を満たす (結線する prioritypass.NewConnectHandler がコンパイル時に確かめる)
type Publisher struct {
	publisher *gcppubsub.Publisher
}

// NewPublisher は申込を送る publisher を組み立てる
//
// Publisher は publish のたびに作り直さず使い回す、トピックごとに batching の goroutine を持つため
func NewPublisher(client *gcppubsub.Client) *Publisher {
	return &Publisher{publisher: client.Publisher(requestedTopicID)}
}

func (p *Publisher) PublishRequested(ctx context.Context, pass *prioritypassmodel.PriorityPass) error {
	message, err := newRequestedMessage(pass)
	if err != nil {
		return err
	}

	result := p.publisher.Publish(ctx, message)

	// publishedAt を書いてよいかの判断に使うため、送信が確かめられるまでここで待つ
	if _, err := result.Get(ctx); err != nil {
		return fmt.Errorf("publish priority pass %q: %w",
			pass.ID(),
			err,
		)
	}

	return nil
}

// newRequestedMessage は送る本文と属性を組み立てる
//
// 購読側 (#116) が読む形そのものなので、Publish から切り離して単体で確かめられるようにしている
func newRequestedMessage(pass *prioritypassmodel.PriorityPass) (*gcppubsub.Message, error) {
	data, err := json.Marshal(requestedMessage{
		PassID:       pass.ID().String(),
		ParkID:       pass.ParkID().String(),
		TicketID:     pass.TicketID().String(),
		AttractionID: pass.AttractionID().String(),
		TimeSlotID:   pass.TimeSlotID().String(),
		RequestedAt:  pass.CreatedAt(),
	})
	if err != nil {
		return nil, fmt.Errorf("encode priority pass %q: %w",
			pass.ID(),
			err,
		)
	}

	return &gcppubsub.Message{
		Data:       data,
		Attributes: map[string]string{attributePassID: pass.ID().String()},
	}, nil
}

// Stop は送り残しを送り切ってから publish の goroutine を止める
func (p *Publisher) Stop() { p.publisher.Stop() }
