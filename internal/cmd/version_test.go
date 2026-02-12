package cmd

import (
	"runtime/debug"
	"testing"
)

func TestGetVersion(t *testing.T) {
	tests := []struct {
		name          string
		buildInfoFunc func() (*debug.BuildInfo, bool)
		want          string
	}{
		{
			name: "returns dev when build info is not available",
			// Simulates: binary built without module support
			buildInfoFunc: func() (*debug.BuildInfo, bool) {
				return nil, false
			},
			want: "dev",
		},
		{
			name: "returns dev when version is (devel)",
			// Simulates: go run ./cmd/code-minions (local build)
			buildInfoFunc: func() (*debug.BuildInfo, bool) {
				return &debug.BuildInfo{
					Main: debug.Module{Version: "(devel)"},
				}, true
			},
			want: "dev",
		},
		{
			name: "returns dev when version is empty",
			buildInfoFunc: func() (*debug.BuildInfo, bool) {
				return &debug.BuildInfo{
					Main: debug.Module{Version: ""},
				}, true
			},
			want: "dev",
		},
		{
			name: "returns module version when set",
			// Simulates: go install ...@v0.1.0
			buildInfoFunc: func() (*debug.BuildInfo, bool) {
				return &debug.BuildInfo{
					Main: debug.Module{Version: "v0.1.0"},
				}, true
			},
			want: "v0.1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save the original and restore after the test
			original := readBuildInfo
			readBuildInfo = tt.buildInfoFunc
			t.Cleanup(func() { readBuildInfo = original })

			got := getVersion()
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
