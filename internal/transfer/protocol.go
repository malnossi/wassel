package transfer

// HandshakePayload is the metadata sent from sender to receiver over the TCP control channel.
type HandshakePayload struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	DownloadURL string `json:"download_url"`
	Checksum    string `json:"checksum,omitempty"` // SHA-256 hex digest for integrity verification
}

// HandshakeResponse is the receiver's accept/decline reply on the TCP control channel.
type HandshakeResponse struct {
	Accepted bool `json:"accepted"`
}

// PingResponse is the reply to a TCP discovery ping (ID == "ping").
type PingResponse struct {
	DeviceName string `json:"device_name"`
}

// Transfer status constants used across sender and receiver sessions.
const (
	StatusConnecting   = "connecting"
	StatusTransferring = "transferring"
	StatusCompleted    = "completed"
	StatusFailed       = "failed"
	StatusDeclined     = "declined"
	StatusCancelled    = "cancelled"
)
