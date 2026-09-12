package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSplitOrigins(t *testing.T) {
	tests := map[string]struct {
		in   string
		want []string
	}{
		"正常系_1件のときそのまま返ること": {
			in:   "http://localhost:5173",
			want: []string{"http://localhost:5173"},
		},
		"正常系_カンマ区切りを分けること": {
			in:   "http://localhost:5173,https://example.com",
			want: []string{"http://localhost:5173", "https://example.com"},
		},
		"正常系_前後の空白を落とすこと": {
			in:   " http://localhost:5173 , https://example.com ",
			want: []string{"http://localhost:5173", "https://example.com"},
		},
		"境界値_空文字のとき何も返さないこと": {
			in:   "",
			want: nil,
		},
		"境界値_カンマだけのとき何も返さないこと": {
			in:   ",,",
			want: nil,
		},
		"境界値_末尾のカンマを許可対象にしないこと": {
			in:   "http://localhost:5173,",
			want: []string{"http://localhost:5173"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := splitOrigins(tt.in)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("splitOrigins(%q) の差分 (-want +got):\n%s", tt.in, diff)
			}
		})
	}
}
