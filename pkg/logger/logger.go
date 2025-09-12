package logger

import (
    "os"

    "github.com/rs/zerolog"
)

// New creates a zerolog.Logger with the provided level string (e.g., "debug", "info").
func New(level string) zerolog.Logger {
    lvl, err := zerolog.ParseLevel(level)
    if err != nil {
        lvl = zerolog.InfoLevel
    }

    zerolog.SetGlobalLevel(lvl)
    logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
    return logger
}


