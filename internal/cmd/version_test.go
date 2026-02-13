package cmd

import (
	"runtime/debug"
	"testing"
)

func TestGetVersion(t *testing.T) {
	tests := []struct {
		name          string
		version       string
		buildInfoFunc func() (*debug.BuildInfo, bool)
		want          string
	}{
		{
			name:    "returns ldflags version when set",
			version: "v0.5.0",
			// BuildInfo is irrelevant when ldflags version is set
			buildInfoFunc: func() (*debug.BuildInfo, bool) {
				return nil, false
			},
			want: "v0.5.0",
		},
		{
			name:    "ldflags version takes precedence over build info",
			version: "v1.0.0",
			buildInfoFunc: func() (*debug.BuildInfo, bool) {
				return &debug.BuildInfo{
					Main: debug.Module{Version: "v0.9.0"},
				}, true
			},
			want: "v1.0.0",
		},
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
			// Save the originals and restore after the test
			originalBuildInfo := readBuildInfo
			originalVersion := Version
			readBuildInfo = tt.buildInfoFunc
			Version = tt.version
			t.Cleanup(func() {
				readBuildInfo = originalBuildInfo
				Version = originalVersion
			})

			got := getVersion()
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
