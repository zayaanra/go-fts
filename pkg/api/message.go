package api

const (
	INITIAL_CONNECT       = 0
	CONFIRMATION          = 1
	SHARE_CONNECTION_INFO = 2
	SHARE_PUBLIC_KEY	  = 4
	SHARE_FILE_DATA       = 5
)

type Message struct {
	Protocol int
	PublicKey []byte
	SessionID string
	Data       []byte
}

type Peer interface {
	Start() error
	Listen() error
	HandlePublicKeyExchange([]byte) ([]byte, error)
	HandleIPExchange([]byte) error
	Close() error
}