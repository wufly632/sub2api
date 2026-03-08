//go:build unit

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateServiceCompareVersions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		current string
		latest  string
		want    int
	}{
		{
			name:    "four_segment_latest_newer",
			current: "0.1.91.0",
			latest:  "0.1.91.2",
			want:    -1,
		},
		{
			name:    "four_segment_current_newer",
			current: "0.1.91.3",
			latest:  "0.1.91.2",
			want:    1,
		},
		{
			name:    "missing_segment_treated_as_zero",
			current: "0.1.91",
			latest:  "0.1.91.0",
			want:    0,
		},
		{
			name:    "v_prefix_supported",
			current: "v0.1.91.0",
			latest:  "v0.1.92.0",
			want:    -1,
		},
		{
			name:    "same_version",
			current: "0.1.91.2",
			latest:  "0.1.91.2",
			want:    0,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, compareVersions(tc.current, tc.latest))
		})
	}
}
