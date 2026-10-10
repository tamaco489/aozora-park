package pubsubpush

import "errors"

// errMissingMessageID はエンベロープに message が無いことを示す
//
// 本文が空のメッセージは機能側が扱えるため、エンベロープの欠落とは区別する
var errMissingMessageID = errors.New("pubsubpush: message.messageId is empty")
