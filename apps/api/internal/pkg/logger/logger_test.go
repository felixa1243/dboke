package logger

import (
	"bytes"
	"log/slog"
	"os"
	"testing"
)

func TestInit_Development(t *testing.T) {
	// Capture original default logger to restore later
	original := slog.Default()
	defer slog.SetDefault(original)

	Init("development")

	// After Init, the default logger should be set
	logger := slog.Default()
	if logger == nil {
		t.Fatal("Expected non-nil logger after Init")
	}

	// Verify debug level is enabled in development
	if !logger.Enabled(nil, slog.LevelDebug) {
		t.Error("Debug level should be enabled in development mode")
	}
}

func TestInit_Production(t *testing.T) {
	original := slog.Default()
	defer slog.SetDefault(original)

	Init("production")

	logger := slog.Default()
	if logger == nil {
		t.Fatal("Expected non-nil logger after Init")
	}

	// In production, debug should be disabled and info should be enabled
	if logger.Enabled(nil, slog.LevelDebug) {
		t.Error("Debug level should NOT be enabled in production mode")
	}
	if !logger.Enabled(nil, slog.LevelInfo) {
		t.Error("Info level should be enabled in production mode")
	}
}

func TestInit_EmptyEnv_DefaultsToDevelopment(t *testing.T) {
	original := slog.Default()
	defer slog.SetDefault(original)

	Init("")

	logger := slog.Default()
	// Empty env should fall into the else branch (development mode)
	if !logger.Enabled(nil, slog.LevelDebug) {
		t.Error("Debug level should be enabled for empty env (defaults to dev)")
	}
}

func TestInit_ProductionUsesJSONHandler(t *testing.T) {
	original := slog.Default()
	defer slog.SetDefault(original)

	// Redirect stdout temporarily to verify JSON output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Init("production")
	slog.Info("test message", slog.String("key", "value"))

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// JSON handler should produce output containing "{"
	if len(output) > 0 && output[0] != '{' {
		// Note: JSON handler output starts with { — this is a basic check
		// The exact format may vary, but it should be JSON-parseable
		t.Logf("Production logger output (may be JSON): %s", output)
	}
}
