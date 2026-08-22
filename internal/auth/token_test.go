package auth

import (
	"testing"
	"time"
)

func TestTokenRoundTripAndExpiry(t *testing.T) {
	manager := NewManager("0123456789abcdef0123456789abcdef", time.Hour)
	now := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	manager.now = func() time.Time { return now }
	token, err := manager.Issue(42, "customer")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := manager.Parse(token)
	if err != nil || claims.UserID != 42 {
		t.Fatalf("claims=%+v err=%v", claims, err)
	}
	manager.now = func() time.Time { return now.Add(2 * time.Hour) }
	if _, err := manager.Parse(token); err == nil {
		t.Fatal("expired token accepted")
	}
}
