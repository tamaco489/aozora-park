// Package apperr はアプリケーションのエラーの語彙と connect のコードへの変換を持つ
package apperr

import (
	"connectrpc.com/connect"
)

// Kind はエラーの分類
//
// 値を増やせるのはこのパッケージだけにする
// 機能パッケージが増やせるのは Code の文字列だけで、分類は共通の語彙に揃える
type Kind string

const (
	KindInvalidArgument Kind = "invalid_argument"
	KindNotFound        Kind = "not_found"
	KindConflict        Kind = "conflict"
	KindUnavailable     Kind = "unavailable"
)

// ConnectCode は Kind に対応する connect のコードを返す
func (k Kind) ConnectCode() connect.Code {
	switch k {
	case KindInvalidArgument:
		return connect.CodeInvalidArgument
	case KindNotFound:
		return connect.CodeNotFound
	case KindConflict:
		return connect.CodeFailedPrecondition
	case KindUnavailable:
		return connect.CodeUnavailable
	default:
		return connect.CodeInternal
	}
}

// Error は機能パッケージが定義するエラー
type Error struct {
	Kind    Kind
	Code    string // API とログに出す機械可読なコード
	Message string
}

// New はセンチネルエラーを作る
//
// 機能パッケージの domain/model/errors.go から呼ぶ
func New(kind Kind, code, message string) *Error {
	return &Error{
		Kind:    kind,
		Code:    code,
		Message: message,
	}
}

func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}
