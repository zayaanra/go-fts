package api

const (
	CONNECT          = 0
	ACKNOWLEDGE      = 1
	SHARE_PUBLIC_KEY = 2
	SHARE_IP         = 3
	SHARE_FILE_DATA  = 4
)

// TODO: Make message handling more structured and suitable for transmission
type Message struct {
	Protocol  int
	PublicKey []byte
	SessionID string
	Data      []byte
}

type File struct {
	Length []byte
	Data   []byte
}
