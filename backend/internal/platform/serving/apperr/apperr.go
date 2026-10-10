// Package apperr はアプリケーションのエラーの語彙と connect のコードへの変換を持つ
package apperr

import (
	"errors"

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
	Retry   bool // Retry は同じ入力でやり直す価値があるか
}

// New は再実行しても結果が変わらない失敗のセンチネルを生成する
//
// 機能パッケージの domain/model/errors.go から呼ぶ
func New(kind Kind, code, message string) *Error {
	return &Error{
		Kind:    kind,
		Code:    code,
		Message: message,
		Retry:   false,
	}
}

// NewRetryable は再実行で直りうる失敗のセンチネルを生成する
//
// 時間をおけば先行する処理が追いつく場合に使う (枠の生成を待つなど)
// Kind からは決められないため、センチネルを定義する側が選ぶ
func NewRetryable(kind Kind, code, message string) *Error {
	return &Error{
		Kind:    kind,
		Code:    code,
		Message: message,
		Retry:   true,
	}
}

// Retryable は再実行で直りうる失敗かを返す
//
// 分類していないエラーは true にする
// SDK やネットワークの失敗がここに入り、握りつぶすより再配信させるほうが安全なため
func Retryable(err error) bool {
	if err == nil {
		return false
	}

	// %w で包まれていても中身まで辿る (infrastructure は文脈を足して返すため)
	// 返る値はセンチネルを定義したときに決まる (New なら false, NewRetryable なら true)
	if appErr, ok := errors.AsType[*Error](err); ok {
		return appErr.Retry
	}

	return true
}

func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}
