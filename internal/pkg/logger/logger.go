package loclog

import (
	"log/slog"
	"os"
)

// ToDo Добавить нормальный логгер, Zap например
func Init() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
