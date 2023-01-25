package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseID(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		out    int
		errStr string
	}{
		{
			name:   "not ok - alpha characters in ID",
			url:    "a/b/123a",
			errStr: "atoi: strconv.Atoi: parsing \"123a\": invalid syntax",
		},
		{
			name: "ok",
			url:  "a/b/123",
			out:  123,
		},
	}

	for _, tt := range tests {
		out, err := parseID(tt.url)
		if tt.errStr != "" {
			require.EqualError(t, err, tt.errStr)
		} else {
			require.Nil(t, err)
		}
		require.Equal(t, tt.out, out)
	}
}
