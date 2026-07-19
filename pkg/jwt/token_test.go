package jwt

import (
	"strings"
	"testing"
	"time"
)

const testSecret = "0123456789abcdef0123456789abcdef"

func TestManagerGenerateAndVerify(t *testing.T) {
	manager, err := NewManager(testSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	token, err := manager.GenerateToken(42, "alice", "alice@example.com")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := manager.VerifyToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != 42 || claims.Username != "alice" || claims.Email != "alice@example.com" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestManagerRejectsInvalidConfigurationAndTokens(t *testing.T) {
	if _, err := NewManager("short", time.Hour); err == nil {
		t.Fatal("expected short secret to be rejected")
	}
	if _, err := NewManager(testSecret, 0); err == nil {
		t.Fatal("expected non-positive TTL to be rejected")
	}

	manager, err := NewManager(testSecret, time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	token, err := manager.GenerateToken(1, "alice", "alice@example.com")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Millisecond)
	if _, err := manager.VerifyToken(token); err == nil {
		t.Fatal("expected expired token to be rejected")
	}

	other, err := NewManager(strings.Repeat("x", 32), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	validToken, err := other.GenerateToken(1, "alice", "alice@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.VerifyToken(validToken); err == nil {
		t.Fatal("expected token signed with another secret to be rejected")
	}
}
