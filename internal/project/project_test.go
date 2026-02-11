package project

import "testing"

func TestParseRemoteURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		wantHost   string
		wantPath   string
		shouldFail bool
	}{
		{
			name:     "ssh scp style",
			input:    "git@gitlab.com:group/sub/repo.git",
			wantHost: "gitlab.com",
			wantPath: "group/sub/repo",
		},
		{
			name:     "https style",
			input:    "https://gitlab.com/group/sub/repo.git",
			wantHost: "gitlab.com",
			wantPath: "group/sub/repo",
		},
		{
			name:     "ssh url style",
			input:    "ssh://git@gitlab.example.com/group/sub/repo.git",
			wantHost: "gitlab.example.com",
			wantPath: "group/sub/repo",
		},
		{
			name:       "invalid",
			input:      "file:///tmp/repo",
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			host, path, err := ParseRemoteURL(tt.input)
			if tt.shouldFail {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseRemoteURL() error = %v", err)
			}
			if host != tt.wantHost {
				t.Fatalf("host = %q want %q", host, tt.wantHost)
			}
			if path != tt.wantPath {
				t.Fatalf("path = %q want %q", path, tt.wantPath)
			}
		})
	}
}
