package transfer

import (
	"io"
	"time"
)

// ProgressWriter wraps an io.Writer and emits throttled progress callbacks.
// The throttle interval prevents IPC event flooding during high-speed transfers.
type ProgressWriter struct {
	w           io.Writer
	total       int64
	current     int64
	lastUpdate  time.Time
	throttleDur time.Duration
	onProgress  func(current int64, total int64)
}

// NewProgressWriter creates a progress-tracking writer that throttles callbacks.
func NewProgressWriter(w io.Writer, total int64, throttleDur time.Duration, onProgress func(current int64, total int64)) *ProgressWriter {
	return &ProgressWriter{
		w:           w,
		total:       total,
		throttleDur: throttleDur,
		onProgress:  onProgress,
		lastUpdate:  time.Time{}, // Force immediate update on first write
	}
}

func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.w.Write(p)
	if n > 0 {
		pw.current += int64(n)
		now := time.Now()
		// Emit progress update if throttle duration has elapsed or we reached 100% completion
		if now.Sub(pw.lastUpdate) >= pw.throttleDur || pw.current == pw.total {
			pw.lastUpdate = now
			pw.onProgress(pw.current, pw.total)
		}
	}
	return
}
