package render

import "testing"

func Test_joinQuotedStrings(t *testing.T) {
	tests := []struct {
		name string
		strs []string
		want string
	}{
		{
			name: "simple",
			strs: []string{"foo bar", "bar baz"},
			want: `"foo bar" "bar baz"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := joinQuotedStrings(tt.strs); got != tt.want {
				t.Errorf("joinQuotedStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}
