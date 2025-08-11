package sso

import (
	"testing"
)

func TestParseSessionFromJSON(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		wantSession *SSOSession
		wantErr     bool
	}{
		{
			name: "valid session",
			json: `{
				"startUrl": "https://example.awsapps.com/start",
				"region": "us-east-1",
				"accessToken": "eyJhbGciOiJ...",
				"expiresAt": "2024-12-31T23:59:59Z"
			}`,
			wantSession: &SSOSession{
				StartUrl:    "https://example.awsapps.com/start",
				Region:      "us-east-1",
				AccessToken: "eyJhbGciOiJ...",
				ExpiresAt:   "2024-12-31T23:59:59Z",
			},
			wantErr: false,
		},
		{
			name:        "invalid JSON",
			json:        `{invalid json}`,
			wantSession: nil,
			wantErr:     true,
		},
		{
			name:        "empty JSON object",
			json:        `{}`,
			wantSession: &SSOSession{},
			wantErr:     false,
		},
		{
			name:        "empty string",
			json:        ``,
			wantSession: nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := ParseSessionFromJSON([]byte(tt.json))

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSessionFromJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantSession == nil {
				if session != nil {
					t.Errorf("ParseSessionFromJSON() = %+v, want nil", session)
				}
			} else {
				if session == nil {
					t.Errorf("ParseSessionFromJSON() = nil, want %+v", tt.wantSession)
				} else if *session != *tt.wantSession {
					t.Errorf("ParseSessionFromJSON() = %+v, want %+v", session, tt.wantSession)
				}
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt string
		want      bool
	}{
		{
			name:      "valid future date",
			expiresAt: "2999-12-31T23:59:59Z",
			want:      true,
		},
		{
			name:      "valid past date",
			expiresAt: "2020-01-01T00:00:00Z",
			want:      false,
		},
		{
			name:      "RFC3339 with timezone offset",
			expiresAt: "2030-12-31T23:59:59+09:00",
			want:      true,
		},
		{
			name:      "empty string",
			expiresAt: "",
			want:      false,
		},
		{
			name:      "invalid format",
			expiresAt: "not-a-date",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &SSOSession{
				ExpiresAt: tt.expiresAt,
			}

			if got := session.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
