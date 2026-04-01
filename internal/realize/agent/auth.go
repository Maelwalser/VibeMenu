package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// claudeCredentials mirrors the relevant fields of ~/.claude/.credentials.json.
type claudeCredentials struct {
	ClaudeAiOauth struct {
		AccessToken string `json:"accessToken"`
		ExpiresAt   int64  `json:"expiresAt"` // milliseconds since epoch
	} `json:"claudeAiOauth"`
}

// ResolveAuthToken returns the best available auth token using this priority:
//  1. explicit token (from CLI flag or env var)
//  2. Claude Code OAuth token from ~/.claude/.credentials.json (if valid)
//  3. empty string → SDK falls back to ANTHROPIC_API_KEY
func ResolveAuthToken(explicit string) (token string, source string) {
	if explicit != "" {
		return explicit, "explicit"
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", "api-key"
	}

	data, err := os.ReadFile(filepath.Join(home, ".claude", ".credentials.json"))
	if err != nil {
		return "", "api-key"
	}

	var creds claudeCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", "api-key"
	}

	tok := creds.ClaudeAiOauth.AccessToken
	exp := creds.ClaudeAiOauth.ExpiresAt
	if tok == "" {
		return "", "api-key"
	}

	// Reject tokens that expire in less than 60 seconds.
	if exp > 0 && time.Now().UnixMilli() > exp-60_000 {
		return "", "api-key (Claude Code OAuth token expired)"
	}

	return tok, "Claude Code OAuth token (~/.claude/.credentials.json)"
}
