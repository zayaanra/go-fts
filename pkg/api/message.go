package api

const (
	CONNECT = 0
	ACKNOWLEDGE = 1
	SHARE_CONNECTION_INFO = 2
	SHARE_PUBLIC_KEY = 4
	SHARE_FILE_DATA = 5
)

type Message struct {
	Protocol int
	PublicKey []byte
	SessionID string
	Data       []byte
}