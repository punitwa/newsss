package logger

import (
    "sync"
    "time"

    "github.com/rs/zerolog"
)

// LogRotator is a lightweight placeholder that periodically logs a heartbeat.
// It can be extended to implement size/time-based rotation if needed.
type LogRotator struct {
    logger zerolog.Logger
    files  []string

    mu     sync.Mutex
    ticker *time.Ticker
    stopCh chan struct{}
    running bool
}

// NewLogRotator creates a new LogRotator for the provided files.
func NewLogRotator(logger zerolog.Logger, files []string) *LogRotator {
    return &LogRotator{
        logger: logger.With().Str("component", "log_rotator").Logger(),
        files:  files,
        stopCh: make(chan struct{}),
    }
}

// Start begins the rotation loop (currently a harmless heartbeat to keep parity with callers).
func (lr *LogRotator) Start() {
    lr.mu.Lock()
    defer lr.mu.Unlock()
    if lr.running {
        return
    }
    lr.ticker = time.NewTicker(6 * time.Hour)
    lr.running = true

    go func() {
        lr.logger.Debug().Strs("files", lr.files).Msg("LogRotator started")
        for {
            select {
            case <-lr.ticker.C:
                // Placeholder: rotate or truncate when implementing real policy
                lr.logger.Debug().Int("file_count", len(lr.files)).Msg("LogRotator tick")
            case <-lr.stopCh:
                lr.logger.Debug().Msg("LogRotator stopping")
                return
            }
        }
    }()
}

// Stop stops the rotation loop.
func (lr *LogRotator) Stop() {
    lr.mu.Lock()
    defer lr.mu.Unlock()
    if !lr.running {
        return
    }
    lr.running = false
    if lr.ticker != nil {
        lr.ticker.Stop()
    }
    close(lr.stopCh)
}

// GetLogFiles returns the known log files for rotation.
// This is a simple helper aligned with how the application writes logs today.
func GetLogFiles() []string {
    return []string{
        "api-gateway.log",
        "collector.log",
        "processor.log",
        "frontend.log",
        "react-app.log",
    }
}


