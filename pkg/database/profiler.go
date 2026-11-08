package database

import (
	"sync"
	"time"
)

// ProfiledQuery represents an executed SQL query with profiling metadata.
type ProfiledQuery struct {
	SQL        string        `json:"sql"`
	Bindings   []any         `json:"bindings"`
	Duration   time.Duration `json:"duration"`
	DurationMS float64       `json:"duration_ms"`
	Timestamp  time.Time     `json:"timestamp"`
	Error      string        `json:"error,omitempty"`
}

// Profiler records executed database queries in a thread-safe circular buffer.
type Profiler struct {
	queries []ProfiledQuery
	max     int
	mu      sync.RWMutex
	enabled bool
}

var (
	globalProfiler *Profiler
	profilerOnce   sync.Once
)

// GetProfiler returns the singleton global query profiler.
func GetProfiler() *Profiler {
	profilerOnce.Do(func() {
		globalProfiler = &Profiler{
			queries: make([]ProfiledQuery, 0, 100),
			max:     100,
			enabled: true,
		}
	})
	return globalProfiler
}

// Record logs an executed SQL query with execution duration.
func (p *Profiler) Record(sqlQuery string, bindings []any, duration time.Duration, err error) {
	if !p.enabled {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	errStr := ""
	if err != nil {
		errStr = err.Error()
	}

	pq := ProfiledQuery{
		SQL:        sqlQuery,
		Bindings:   bindings,
		Duration:   duration,
		DurationMS: float64(duration.Microseconds()) / 1000.0,
		Timestamp:  time.Now(),
		Error:      errStr,
	}

	if len(p.queries) >= p.max {
		p.queries = append(p.queries[1:], pq)
	} else {
		p.queries = append(p.queries, pq)
	}
}

// Queries returns a copy of all recorded queries in reverse chronological order (newest first).
func (p *Profiler) Queries() []ProfiledQuery {
	p.mu.RLock()
	defer p.mu.RUnlock()
	res := make([]ProfiledQuery, len(p.queries))
	for i, q := range p.queries {
		res[len(p.queries)-1-i] = q
	}
	return res
}

// Clear resets recorded queries.
func (p *Profiler) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.queries = p.queries[:0]
}

// Stats returns summary statistics: total queries and total execution time in milliseconds.
func (p *Profiler) Stats() (int, float64) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var totalMS float64
	for _, q := range p.queries {
		totalMS += q.DurationMS
	}
	return len(p.queries), totalMS
}

