package toolbox

import (
	"log/slog"
	"os"
)

var (
	Logger 	   *slog.Logger
)

// Init logger
func InitLogger(lvl string) {
    logLevel := new(slog.LevelVar)
    // Set log level
    switch lvl {
    case "debug":
        logLevel.Set(slog.LevelDebug)
    case "info":
        logLevel.Set(slog.LevelInfo)
    case "error":
        logLevel.Set(slog.LevelError)
    // Default is warn
    default:
        logLevel.Set(slog.LevelWarn)
    }

    // Create logger
    Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
        Level: logLevel,
    }))

    Logger.Debug("successfully created logger")
}