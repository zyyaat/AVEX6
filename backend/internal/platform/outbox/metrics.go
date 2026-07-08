// Package outbox metrics: lightweight metrics for the outbox publisher worker.
//
// In Phase 2, these are simple counters logged via slog. When the system
// module is implemented, these will be upgraded to OpenTelemetry metrics
// (Prometheus-compatible via the OTel SDK).
package outbox

import (
	"sync/atomic"
)

// Metrics tracks outbox publisher worker statistics.
type Metrics struct {
	PublishedCount atomic.Int64
	FailedCount    atomic.Int64
	TotalProcessed atomic.Int64
	LastBatchSize  atomic.Int64
}

// NewMetrics creates a new Metrics instance.
func NewMetrics() *Metrics {
	return &Metrics{}
}

// RecordPublished increments the published counter.
func (m *Metrics) RecordPublished() {
	m.PublishedCount.Add(1)
	m.TotalProcessed.Add(1)
}

// RecordFailed increments the failed counter.
func (m *Metrics) RecordFailed() {
	m.FailedCount.Add(1)
	m.TotalProcessed.Add(1)
}

// SetLastBatchSize records the size of the last fetched batch.
func (m *Metrics) SetLastBatchSize(n int) {
	m.LastBatchSize.Store(int64(n))
}

// Snapshot returns a point-in-time copy of the metrics.
type MetricsSnapshot struct {
	PublishedCount int64
	FailedCount    int64
	TotalProcessed int64
	LastBatchSize  int64
}

// Snapshot returns current metric values.
func (m *Metrics) Snapshot() MetricsSnapshot {
	return MetricsSnapshot{
		PublishedCount: m.PublishedCount.Load(),
		FailedCount:    m.FailedCount.Load(),
		TotalProcessed: m.TotalProcessed.Load(),
		LastBatchSize:  m.LastBatchSize.Load(),
	}
}
