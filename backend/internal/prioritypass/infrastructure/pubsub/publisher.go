// Package pubsub は優先パスの出来事を Pub/Sub へ送り出す
package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	gcppubsub "cloud.google.com/go/pubsub/v2"

	prioritypassmodel "github.com/tamaco489/aozora-park/backend/internal/prioritypass/domain/model"
)

// requestedTopicID は申込を流すトピック
//
// infra/modules/pubsub/main.tf の google_pubsub_topic.prioritypass_requested と同じ値にする
const requestedTopicID = "prioritypass.requested"

// requestedMessage は prioritypass.requested に流す本文
//
//   - 申込の内容は Firestore を正とし、ここには識別子だけを載せる
//   - 項目を足すと購読側が読まなくても契約が増えるため、必要になってから足す
type requestedMessage struct {
	PassID string `json:"passId"`
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

// newRequestedMessage は送る本文を組み立てる
//
// 購読側が読む形そのものなので、Publish から切り離して単体で確かめられるようにしている
func newRequestedMessage(pass *prioritypassmodel.PriorityPass) (*gcppubsub.Message, error) {
	data, err := json.Marshal(requestedMessage{PassID: pass.ID().String()})
	if err != nil {
		return nil, fmt.Errorf("encode priority pass %q: %w",
			pass.ID(),
			err,
		)
	}

	return &gcppubsub.Message{Data: data}, nil
}

// Stop は送り残しを送り切ってから publish の goroutine を止める
func (p *Publisher) Stop() { p.publisher.Stop() }
