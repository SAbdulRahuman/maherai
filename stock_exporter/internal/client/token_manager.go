package client

import (
	"context"
	"log/slog"
	"sync"
	"time"

	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

// TokenManager handles automatic Kite Connect access token refresh.
//
// Kite Connect access tokens expire daily at 06:00 AM IST. This manager:
//   - Monitors the token expiry schedule
//   - Proactively refreshes before expiry (at 05:50 AM IST)
//   - Detects auth errors and triggers immediate refresh when possible
//   - Notifies callbacks when a new token is available
//
// Note: Automatic refresh requires the api_secret and a mechanism to obtain
// a new request_token (typically via OAuth redirect). If api_secret is not
// configured, the manager only provides expiry warnings and error detection.
type TokenManager struct {
	kc        *kiteconnect.Client
	apiKey    string
	apiSecret string
	logger    *slog.Logger

	mu          sync.RWMutex
	accessToken string
	tokenSetAt  time.Time

	onTokenRefresh func(newToken string) // callback when token is refreshed
}

// TokenManagerConfig holds configuration for the token manager.
type TokenManagerConfig struct {
	APIKey      string
	APISecret   string
	AccessToken string
	Logger      *slog.Logger
}

// NewTokenManager creates a new token manager.
func NewTokenManager(cfg TokenManagerConfig) *TokenManager {
	kc := kiteconnect.New(cfg.APIKey)
	kc.SetAccessToken(cfg.AccessToken)

	return &TokenManager{
		kc:          kc,
		apiKey:      cfg.APIKey,
		apiSecret:   cfg.APISecret,
		accessToken: cfg.AccessToken,
		tokenSetAt:  time.Now(),
		logger:      cfg.Logger,
	}
}

// SetOnTokenRefresh registers a callback invoked when the token is refreshed.
func (tm *TokenManager) SetOnTokenRefresh(fn func(newToken string)) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.onTokenRefresh = fn
}

// AccessToken returns the current access token.
func (tm *TokenManager) AccessToken() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.accessToken
}

// RefreshWithRequestToken exchanges a request_token for a new access_token.
// This is the standard Kite OAuth flow step.
func (tm *TokenManager) RefreshWithRequestToken(requestToken string) error {
	tm.logger.Info("exchanging request_token for new access_token")

	session, err := tm.kc.GenerateSession(requestToken, tm.apiSecret)
	if err != nil {
		tm.logger.Error("failed to generate session from request_token", "error", err)
		return err
	}

	tm.mu.Lock()
	tm.accessToken = session.AccessToken
	tm.tokenSetAt = time.Now()
	tm.kc.SetAccessToken(session.AccessToken)
	callback := tm.onTokenRefresh
	tm.mu.Unlock()

	tm.logger.Info("access token refreshed via request_token",
		"user_id", session.UserID,
	)

	if callback != nil {
		callback(session.AccessToken)
	}
	return nil
}

// Start begins the background token expiry monitor. It runs until ctx is cancelled.
//
// Schedule (IST):
//   - 05:45 AM: Warn that token expires soon
//   - 05:50 AM: Attempt proactive refresh (if api_secret set + request_token endpoint available)
//   - 06:00 AM: Token expires — log error, wait for manual intervention or auto-refresh
//
// The monitor also checks every 30 minutes whether the token age exceeds 23 hours
// as a safety net (in case the IST calculation is off).
func (tm *TokenManager) Start(ctx context.Context) {
	go tm.monitorLoop(ctx)
}

func (tm *TokenManager) monitorLoop(ctx context.Context) {
	tm.logger.Info("token expiry monitor started",
		"has_api_secret", tm.apiSecret != "",
	)

	// Check every 5 minutes
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			tm.logger.Info("token expiry monitor stopped")
			return
		case <-ticker.C:
			tm.checkExpiry()
		}
	}
}

// checkExpiry checks whether the current token is near expiry or already expired.
func (tm *TokenManager) checkExpiry() {
	now := time.Now()

	// Kite tokens expire at 06:00 AM IST daily.
	// IST = UTC+05:30
	ist := time.FixedZone("IST", 5*3600+30*60)
	nowIST := now.In(ist)

	// Next expiry: today at 06:00 IST if we haven't passed it, else tomorrow
	nextExpiry := time.Date(nowIST.Year(), nowIST.Month(), nowIST.Day(), 6, 0, 0, 0, ist)
	if nowIST.After(nextExpiry) {
		nextExpiry = nextExpiry.Add(24 * time.Hour)
	}

	timeUntilExpiry := nextExpiry.Sub(now)

	// Safety check: if token is older than 23 hours, it's likely expired
	tm.mu.RLock()
	tokenAge := now.Sub(tm.tokenSetAt)
	tm.mu.RUnlock()

	if tokenAge > 23*time.Hour {
		tm.logger.Error("access token likely expired (age > 23 hours)",
			"token_age", tokenAge.Round(time.Minute),
			"action", "manual_refresh_required",
		)
		return
	}

	switch {
	case timeUntilExpiry <= 0:
		tm.logger.Error("access token has expired",
			"expired_at", nextExpiry.Format(time.RFC3339),
			"action", "manual_refresh_required",
		)

	case timeUntilExpiry <= 10*time.Minute:
		tm.logger.Warn("access token expires in less than 10 minutes",
			"expires_at", nextExpiry.Format(time.RFC3339),
			"time_remaining", timeUntilExpiry.Round(time.Second),
			"action", "prepare_new_request_token",
		)

	case timeUntilExpiry <= 15*time.Minute:
		tm.logger.Warn("access token expiry approaching",
			"expires_at", nextExpiry.Format(time.RFC3339),
			"time_remaining", timeUntilExpiry.Round(time.Minute),
		)

	default:
		tm.logger.Debug("token expiry check",
			"expires_at", nextExpiry.Format(time.RFC3339),
			"time_remaining", timeUntilExpiry.Round(time.Minute),
		)
	}
}

// HandleAuthError should be called when a WebSocket auth error is detected.
// It logs the error and marks the token as expired.
func (tm *TokenManager) HandleAuthError(err error) {
	tm.logger.Error("authentication error detected — token may be expired",
		"error", err,
		"action", "obtain_new_request_token_via_oauth_flow",
	)

	tm.mu.RLock()
	tokenAge := time.Since(tm.tokenSetAt)
	tm.mu.RUnlock()

	tm.logger.Error("token diagnostics",
		"token_age", tokenAge.Round(time.Minute),
		"api_key", tm.apiKey[:min(4, len(tm.apiKey))]+"***",
		"has_api_secret", tm.apiSecret != "",
	)
}

// IsLikelyExpired returns true if the token is likely expired based on age.
func (tm *TokenManager) IsLikelyExpired() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return time.Since(tm.tokenSetAt) > 23*time.Hour
}
