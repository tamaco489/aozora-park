package model

// Allocation は割当を試みた結果
//
// 不変条件を持たない入れ物のためフィールドを公開する
type Allocation struct {
	Pass    *PriorityPass // Pass は割当を試みた優先パス
	Changed bool          // Changed は状態の遷移と枠の減算が起きたか
}
