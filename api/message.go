package api

type MessageType int

const (
	CONNECT MessageType = iota
	ACKNOWLEDGE
	SHARE_PUBLIC_KEY
	SHARE_IP
	SHARE_FILE_DATA
)

type Message struct {
	Protocol  MessageType
	SessionID string
	Data      []byte
}