package logger

import (
	"log/slog"
	"os"
)

// Init sets up the global logger based on the environment
func Init(env string) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		AddSource: true,
	}

	if env == "production" {
		opts.Level = slog.LevelInfo
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		opts.Level = slog.LevelDebug
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}
