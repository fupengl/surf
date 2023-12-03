package surf

import "time"

type Performance struct {
	ResponseTime time.Duration
}

func newPerformance() *Performance {
	return &Performance{}
}

func (p *Performance) recordResponseTime(startTime time.Time) {
	p.ResponseTime = time.Since(startTime)
}
