package middleware

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nomdb/backend/internal/logger"
)

// Metrics holds application metrics
type Metrics struct {
	TotalRequests     uint64
	TotalErrors       uint64
	RequestsByMethod  map[string]*uint64
	RequestsByPath    map[string]*uint64
	RequestsByStatus  map[int]*uint64
	ResponseTimes     []time.Duration
	mu                sync.RWMutex
	lastLogTime       time.Time
}

var (
	appMetrics *Metrics
	once       sync.Once
)

// GetMetrics returns the singleton metrics instance
func GetMetrics() *Metrics {
	once.Do(func() {
		appMetrics = &Metrics{
			RequestsByMethod: make(map[string]*uint64),
			RequestsByPath:   make(map[string]*uint64),
			RequestsByStatus: make(map[int]*uint64),
			ResponseTimes:    make([]time.Duration, 0, 1000),
			lastLogTime:      time.Now(),
		}

		// Start periodic metrics logging
		go appMetrics.logPeriodically(5 * time.Minute)
	})
	return appMetrics
}

// RecordRequest records metrics for an HTTP request
func (m *Metrics) RecordRequest(method, path string, status int, duration time.Duration) {
	// Increment total requests
	atomic.AddUint64(&m.TotalRequests, 1)

	// Increment errors if status >= 500
	if status >= 500 {
		atomic.AddUint64(&m.TotalErrors, 1)
	}

	// Record by method
	m.mu.Lock()
	if m.RequestsByMethod[method] == nil {
		var count uint64
		m.RequestsByMethod[method] = &count
	}
	atomic.AddUint64(m.RequestsByMethod[method], 1)

	// Record by path (limit to prevent memory bloat)
	if len(m.RequestsByPath) < 100 {
		if m.RequestsByPath[path] == nil {
			var count uint64
			m.RequestsByPath[path] = &count
		}
		atomic.AddUint64(m.RequestsByPath[path], 1)
	}

	// Record by status
	if m.RequestsByStatus[status] == nil {
		var count uint64
		m.RequestsByStatus[status] = &count
	}
	atomic.AddUint64(m.RequestsByStatus[status], 1)

	// Record response time (limit array size)
	if len(m.ResponseTimes) < 1000 {
		m.ResponseTimes = append(m.ResponseTimes, duration)
	}
	m.mu.Unlock()
}

// GetStats returns current metrics statistics
func (m *Metrics) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Calculate average response time
	var avgDuration time.Duration
	if len(m.ResponseTimes) > 0 {
		var total time.Duration
		for _, d := range m.ResponseTimes {
			total += d
		}
		avgDuration = total / time.Duration(len(m.ResponseTimes))
	}

	// Calculate percentiles (p50, p95, p99)
	p50, p95, p99 := m.calculatePercentiles()

	// Copy maps to avoid race conditions
	byMethod := make(map[string]uint64)
	for k, v := range m.RequestsByMethod {
		byMethod[k] = atomic.LoadUint64(v)
	}

	byPath := make(map[string]uint64)
	for k, v := range m.RequestsByPath {
		byPath[k] = atomic.LoadUint64(v)
	}

	byStatus := make(map[int]uint64)
	for k, v := range m.RequestsByStatus {
		byStatus[k] = atomic.LoadUint64(v)
	}

	return map[string]interface{}{
		"total_requests":   atomic.LoadUint64(&m.TotalRequests),
		"total_errors":     atomic.LoadUint64(&m.TotalErrors),
		"requests_by_method": byMethod,
		"requests_by_path": byPath,
		"requests_by_status": byStatus,
		"avg_response_time": avgDuration.String(),
		"p50_response_time": p50.String(),
		"p95_response_time": p95.String(),
		"p99_response_time": p99.String(),
		"uptime":          time.Since(m.lastLogTime).String(),
	}
}

// calculatePercentiles calculates response time percentiles
func (m *Metrics) calculatePercentiles() (p50, p95, p99 time.Duration) {
	if len(m.ResponseTimes) == 0 {
		return
	}

	// Copy and sort response times for percentile calculation
	sorted := make([]time.Duration, len(m.ResponseTimes))
	copy(sorted, m.ResponseTimes)

	// Use efficient O(n log n) sort
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	p50Index := int(float64(len(sorted)) * 0.50)
	p95Index := int(float64(len(sorted)) * 0.95)
	p99Index := int(float64(len(sorted)) * 0.99)

	if p50Index < len(sorted) {
		p50 = sorted[p50Index]
	}
	if p95Index < len(sorted) {
		p95 = sorted[p95Index]
	}
	if p99Index < len(sorted) {
		p99 = sorted[p99Index]
	}

	return
}

// logPeriodically logs metrics at regular intervals
func (m *Metrics) logPeriodically(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		stats := m.GetStats()
		logger.InfoWithFields("📊 Metrics Summary", stats)
	}
}

// Reset resets all metrics (useful for testing)
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.StoreUint64(&m.TotalRequests, 0)
	atomic.StoreUint64(&m.TotalErrors, 0)
	m.RequestsByMethod = make(map[string]*uint64)
	m.RequestsByPath = make(map[string]*uint64)
	m.RequestsByStatus = make(map[int]*uint64)
	m.ResponseTimes = make([]time.Duration, 0, 1000)
	m.lastLogTime = time.Now()
}
