package auth

import (
	"testing"
	"time"
)

func TestTokensIssueAndParseRefresh(t *testing.T) {
	tokens, err := NewTokens(Config{Secret: "unit-test-secret", Issuer: "msg-push"})
	if err != nil {
		t.Fatalf("new tokens: %v", err)
	}

	pair, err := tokens.Issue(42, "admin")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("issue returned empty token")
	}
	if !pair.ExpiresAt.After(time.Now()) {
		t.Fatalf("expires_at = %v, want in the future", pair.ExpiresAt)
	}

	id, err := tokens.ParseRefresh(pair.RefreshToken)
	if err != nil {
		t.Fatalf("parse refresh token: %v", err)
	}
	if id != 42 {
		t.Fatalf("subject = %d, want 42", id)
	}

	// 访问令牌不能当作刷新令牌使用
	if _, err := tokens.ParseRefresh(pair.AccessToken); err == nil {
		t.Fatal("access token accepted as refresh token")
	}
}

func TestNewTokensRejectsEmptySecret(t *testing.T) {
	if _, err := NewTokens(Config{}); err == nil {
		t.Fatal("empty secret accepted")
	}
}
