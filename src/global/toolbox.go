package global

import (
	"log/slog"
	"os"
)

var (
	Logger * slog.Logger
)

func CreateLogger(logLvl *string) {

	// Set log level, default is 'warn'
	logLevel := new(slog.LevelVar)
	switch(*logLvl) {
	case "debug":
		logLevel.Set(slog.LevelDebug)
	case "info":
		logLevel.Set(slog.LevelInfo)
	case "error":
		logLevel.Set(slog.LevelError)
	default:
		logLevel.Set(slog.LevelWarn)
	}

	Logger = slog.New(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: logLevel,
		},
	))

	Logger.Debug("Finished creating custom logger")
}