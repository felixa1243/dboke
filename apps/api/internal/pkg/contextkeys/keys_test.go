package contextkeys

import (
	"context"
	"testing"
)

func TestContextKeys_Uniqueness(t *testing.T) {
	// Ensure all keys are distinct
	keys := []ContextKey{UserIDKey, RoleKey, SessionIDKey}
	seen := make(map[ContextKey]bool)
	for _, k := range keys {
		if seen[k] {
			t.Errorf("Duplicate context key found: %s", k)
		}
		seen[k] = true
	}
}

func TestContextKeys_Values(t *testing.T) {
	if UserIDKey != "user_id" {
		t.Errorf("UserIDKey = %q, want %q", UserIDKey, "user_id")
	}
	if RoleKey != "role" {
		t.Errorf("RoleKey = %q, want %q", RoleKey, "role")
	}
	if SessionIDKey != "session_id" {
		t.Errorf("SessionIDKey = %q, want %q", SessionIDKey, "session_id")
	}
}

func TestContextKeys_StoreAndRetrieve(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, UserIDKey, "user-123")
	ctx = context.WithValue(ctx, RoleKey, "admin")
	ctx = context.WithValue(ctx, SessionIDKey, "session-abc")

	if v := ctx.Value(UserIDKey).(string); v != "user-123" {
		t.Errorf("UserIDKey value = %q, want %q", v, "user-123")
	}
	if v := ctx.Value(RoleKey).(string); v != "admin" {
		t.Errorf("RoleKey value = %q, want %q", v, "admin")
	}
	if v := ctx.Value(SessionIDKey).(string); v != "session-abc" {
		t.Errorf("SessionIDKey value = %q, want %q", v, "session-abc")
	}
}

func TestContextKeys_NoCollisionWithPlainString(t *testing.T) {
	// Ensure ContextKey type doesn't collide with raw string keys
	ctx := context.Background()
	ctx = context.WithValue(ctx, UserIDKey, "from-typed-key")
	ctx = context.WithValue(ctx, "user_id", "from-string-key")

	typedVal := ctx.Value(UserIDKey).(string)
	stringVal := ctx.Value("user_id").(string)

	if typedVal == stringVal {
		t.Error("Typed ContextKey and raw string key produced same context value — type safety failure")
	}
}

func TestContextKeys_MissingKey_ReturnsNil(t *testing.T) {
	ctx := context.Background()
	val := ctx.Value(UserIDKey)
	if val != nil {
		t.Errorf("Expected nil for missing key, got %v", val)
	}
}
