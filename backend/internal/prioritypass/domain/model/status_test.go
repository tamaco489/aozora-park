package model

import (
	"testing"
)

func TestStatusIsValid(t *testing.T) {
	tests := map[string]struct {
		status Status
		want   bool
	}{
		"正常系_requestedの場合_trueになること": {
			status: StatusRequested,
			want:   true,
		},
		"正常系_issuedの場合_trueになること": {
			status: StatusIssued,
			want:   true,
		},
		"正常系_sold_outの場合_trueになること": {
			status: StatusSoldOut,
			want:   true,
		},
		"異常系_却下の場合_falseになること": {
			status: Status("rejected"),
			want:   false,
		},
		"異常系_空の場合_falseになること": {
			status: Status(""),
			want:   false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("Status(%q).IsValid() = %t, want %t",
					tt.status,
					got,
					tt.want,
				)
			}
		})
	}
}
