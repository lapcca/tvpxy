// Package proxy Package constants contains the constants used in the application.
package proxy

const (
	RTP_Payload_MP2T    = 33
	ContentType_MP2T    = "video/MP2T"
	ContentType_DEFAULT = "application/octet-stream"
)

type DATA struct {
	buf []byte
	len int
}
