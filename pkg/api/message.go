package api

const (
	CONNECT = 0
	ACKNOWLEDGE = 1
	SHARE_PUBLIC_KEY = 2
	SHARE_IP = 3
	SHARE_FILE_DATA = 4
)

type Message struct {
	Protocol int
	PublicKey []byte
	SessionID string
	Data       []byte
}