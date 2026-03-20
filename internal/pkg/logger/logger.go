package loclog

import (
	"log/slog"
	"os"
)

func Init(env string) *slog.Logger {
	var logLevel slog.Level
	if env == "development" {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}

	opts := slog.HandlerOptions{
		Level: logLevel,
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &opts))
}
