package sso

import (
	"encoding/json"
	"fmt"
	"time"
)

type SSOSession struct {
	StartUrl    string `json:"startUrl"`
	Region      string `json:"region"`
	AccessToken string `json:"accessToken"`
	ExpiresAt   string `json:"expiresAt"`
}

// ParseSessionFromJSON parses SSO session from JSON data
func ParseSessionFromJSON(data []byte) (*SSOSession, error) {
	var session SSOSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to parse session JSON: %w", err)
	}
	return &session, nil
}

// IsValid checks if the session is still valid
func (s *SSOSession) IsValid() bool {
	expiresAt, err := time.Parse(time.RFC3339, s.ExpiresAt)
	if err != nil {
		return false
	}
	return time.Now().Before(expiresAt)
}
