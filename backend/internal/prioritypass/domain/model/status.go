package model

// Status は優先パスの状態
//
// 却下を表す rejected は、券との突き合わせを行う tickets ができたときに足す
type Status string

const (
	StatusRequested Status = "requested"
	StatusIssued    Status = "issued"
	StatusSoldOut   Status = "sold_out"
)

// IsValid は保存済みの値が既知の状態かを返す
func (s Status) IsValid() bool {
	switch s {
	case StatusRequested, StatusIssued, StatusSoldOut:
		return true
	default:
		return false
	}
}

func (s Status) String() string { return string(s) }
