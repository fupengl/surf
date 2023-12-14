package surf

import (
	"time"
)

// Performance represents the response performance metrics.
type Performance struct {
	clientTrace *clientTrace

	// DNSLookup is a duration that transport took to perform
	DNSLookup time.Duration

	// ConnTime is a duration that took to obtain a successful connection.
	ConnTime time.Duration

	// TCPConnTime is a duration that took to obtain the TCP connection.
	TCPConnTime time.Duration

	// TLSHandshake is a duration that TLS handshake took place.
	TLSHandshake time.Duration

	// ServerTime is a duration that server took to respond first byte.
	ServerTime time.Duration

	// ResponseTime is a duration since first response byte from server to
	ResponseTime time.Duration

	// TotalTime is a duration that total request took end-to-end.
	TotalTime time.Duration

	// IsConnReused is whether this connection has been previously
	IsConnReused bool

	// IsConnWasIdle is whether this connection was obtained from an
	IsConnWasIdle bool

	// ConnIdleTime is a duration how long the connection was previously
	ConnIdleTime time.Duration
}

func (p *Performance) record() {
	ct := p.clientTrace

	ct.endTime = time.Now()
	p.DNSLookup = ct.dnsDone.Sub(ct.dnsStart)
	p.TLSHandshake = ct.tlsHandshakeDone.Sub(ct.tlsHandshakeStart)
	p.ServerTime = ct.gotFirstResponseByte.Sub(ct.gotConn)
	p.IsConnReused = ct.gotConnInfo.Reused
	p.IsConnWasIdle = ct.gotConnInfo.WasIdle
	p.ConnIdleTime = ct.gotConnInfo.IdleTime

	// when connection is reused
	if ct.gotConnInfo.Reused {
		p.TotalTime = ct.endTime.Sub(ct.getConn)
	} else {
		p.TotalTime = ct.endTime.Sub(ct.dnsStart)
	}

	// Only calculate on successful connections
	if !ct.connectDone.IsZero() {
		p.TCPConnTime = ct.connectDone.Sub(ct.dnsDone)
	}

	// Only calculate on successful connections
	if !ct.gotConn.IsZero() {
		p.ConnTime = ct.gotConn.Sub(ct.getConn)
	}

	// Only calculate on successful connections
	if !ct.gotFirstResponseByte.IsZero() {
		p.ResponseTime = ct.endTime.Sub(ct.gotFirstResponseByte)
	}
}
