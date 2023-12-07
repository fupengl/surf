package surf

import "time"

// Performance represents the response performance metrics.
type Performance struct {
	ResponseTime time.Duration // ResponseTime is the duration of the response.
}

// newPerformance creates a new Performance instance.
func newPerformance() *Performance {
	return &Performance{}
}

// recordResponseTime records the duration of the response based on the start time.
func (p *Performance) recordResponseTime(startTime time.Time) {
	p.ResponseTime = time.Since(startTime)
}
