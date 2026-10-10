package apperr

import (
	"errors"
	"fmt"
	"testing"

	"connectrpc.com/connect"
)

func TestKindConnectCode(t *testing.T) {
	tests := map[string]struct {
		kind Kind
		want connect.Code
	}{
		"正常系_invalid_argumentの場合_CodeInvalidArgumentになること": {
			kind: KindInvalidArgument,
			want: connect.CodeInvalidArgument,
		},
		"正常系_not_foundの場合_CodeNotFoundになること": {
			kind: KindNotFound,
			want: connect.CodeNotFound,
		},
		"正常系_conflictの場合_CodeFailedPreconditionになること": {
			kind: KindConflict,
			want: connect.CodeFailedPrecondition,
		},
		"正常系_unavailableの場合_CodeUnavailableになること": {
			kind: KindUnavailable,
			want: connect.CodeUnavailable,
		},
		"異常系_未知のKindの場合_CodeInternalになること": {
			kind: Kind("unknown"),
			want: connect.CodeInternal,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.kind.ConnectCode()
			if got != tt.want {
				t.Errorf("Kind(%q).ConnectCode() = %v, want %v",
					tt.kind,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestErrorError(t *testing.T) {
	err := New(
		KindConflict,
		"PARK_ALREADY_EXISTS",
		"同じパークが既にある",
	)

	want := "PARK_ALREADY_EXISTS: 同じパークが既にある"
	if got := err.Error(); got != want {
		t.Errorf("New(...).Error() = %q, want %q",
			got,
			want,
		)
	}
}

func TestRetryable(t *testing.T) {
	tests := map[string]struct {
		err  error
		want bool
	}{
		"正常系_nilの場合_falseになること": {
			err:  nil,
			want: false,
		},
		"正常系_Newで作った場合_falseになること": {
			err: New(
				KindConflict,
				"PARK_ALREADY_EXISTS",
				"同じパークが既にある",
			),
			want: false,
		},
		"正常系_NewRetryableで作った場合_trueになること": {
			err: NewRetryable(
				KindNotFound,
				"PRIORITY_PASS_TIME_SLOT_NOT_FOUND",
				"枠がまだ無い",
			),
			want: true,
		},
		"正常系_包んだセンチネルの場合_中身の判定に従うこと": {
			err: fmt.Errorf("get time slot: %w", NewRetryable(
				KindNotFound,
				"PRIORITY_PASS_TIME_SLOT_NOT_FOUND",
				"枠がまだ無い",
			)),
			want: true,
		},
		"正常系_分類していないエラーの場合_trueになること": {
			err:  errors.New("firestore unavailable"),
			want: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := Retryable(tt.err)
			if got != tt.want {
				t.Errorf("Retryable(%v) = %t, want %t",
					tt.err,
					got,
					tt.want,
				)
			}
		})
	}
}
