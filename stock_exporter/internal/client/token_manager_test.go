package client

import (
	"testing"
	"time"
)

func TestTokenManager_IsLikelyExpired(t *testing.T) {
	tm := NewTokenManager(TokenManagerConfig{
		APIKey:      "test_key",
		APISecret:   "test_secret",
		AccessToken: "test_token",
		Logger:      nil, // nil is fine for this test
	})

	if tm.IsLikelyExpired() {
		t.Error("fresh token should not be expired")
	}

	// Simulate an old token
	tm.mu.Lock()
	tm.tokenSetAt = time.Now().Add(-24 * time.Hour)
	tm.mu.Unlock()

	if !tm.IsLikelyExpired() {
		t.Error("24-hour-old token should be likely expired")
	}
}

func TestTokenManager_AccessToken(t *testing.T) {
	tm := NewTokenManager(TokenManagerConfig{
		APIKey:      "key",
		APISecret:   "secret",
		AccessToken: "my_token_123",
		Logger:      nil,
	})

	if tm.AccessToken() != "my_token_123" {
		t.Errorf("expected my_token_123, got %q", tm.AccessToken())
	}
}
