package model

import (
	"regexp"
	"unicode/utf8"
)

// timeOfDay は HH:MM の 24 時間表記、proto の制約と同じ値を置く
var timeOfDay = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

// isValidName は表示名が 1 文字から nameMaxLen 文字の範囲にあるかを返す
func isValidName(name string) bool {
	// 文字数で数えるのは、上限が表示の崩れを防ぐための歯止めでバイト数に意味がないため
	l := utf8.RuneCountInString(name)
	return l >= 1 && l <= nameMaxLen
}

// isValidTimeOfDay は HH:MM の 24 時間表記かを返す
func isValidTimeOfDay(s string) bool { return timeOfDay.MatchString(s) }

// isBefore は HH:MM の 2 つの時刻の前後を返す
//
// 桁が揃った 24 時間表記のため、文字列の大小がそのまま時刻の前後になる
func isBefore(a, b string) bool { return a < b }
