// Package collector provides a backward-compatible interface to the refactored collector components.
// This file re-exports the main collector functionality to maintain API compatibility.
package collector

import (
	"news-aggregator/internal/collector/core"
	"news-aggregator/internal/collector/jobs"
	"news-aggregator/internal/collector/scheduling"
	"news-aggregator/internal/collector/sources"
)

// Re-export main interfaces for backward compatibility
type (
	Collector       = core.Collector
	WorkerPool      = core.WorkerPool
	SourceManager   = core.SourceManager
	JobScheduler    = core.JobScheduler
	CollectionJob   = core.CollectionJob
	WorkerPoolStats = core.WorkerPoolStats

	// Configuration and metrics types
	CollectorConfig  = core.CollectorConfig
	CollectorMetrics = core.CollectorMetrics
	Logger           = core.Logger
)

// Re-export constructor functions for backward compatibility
var (
	New                    = core.New
	NewWithConfig          = core.NewWithConfig
	DefaultCollectorConfig = core.DefaultCollectorConfig

	// Component constructors
	NewSourceManager = sources.NewSourceManager
	NewJobScheduler  = scheduling.NewJobScheduler

	// Job constructors
	NewCollectionJob             = jobs.NewCollectionJob
	NewCollectionJobWithPriority = jobs.NewCollectionJobWithPriority
)

// Re-export job-related types and constants
type (
	JobPriority = jobs.JobPriority
	JobStatus   = jobs.JobStatus
	JobResult   = jobs.JobResult
	RetryConfig = jobs.RetryConfig
)

const (
	PriorityLow    = jobs.PriorityLow
	PriorityNormal = jobs.PriorityNormal
	PriorityHigh   = jobs.PriorityHigh
	PriorityUrgent = jobs.PriorityUrgent

	JobStatusPending    = jobs.JobStatusPending
	JobStatusProcessing = jobs.JobStatusProcessing
	JobStatusCompleted  = jobs.JobStatusCompleted
	JobStatusFailed     = jobs.JobStatusFailed
	JobStatusRetrying   = jobs.JobStatusRetrying
)

// Re-export scheduling types
type (
	ScheduleInfo   = scheduling.ScheduleInfo
	SchedulerStats = scheduling.SchedulerStats
)
