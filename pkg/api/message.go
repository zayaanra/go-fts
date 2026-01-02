package api

const (
	INITIAL_CONNECT = 0 // Signifies the message is for sending a passphrase to relay server for PAKE
	CONFIRMATION = 1 // Signifies that both clients (A and B) have connected to the relay server so that the PAKE procss can begin
	SHARE_PBK = 2 // Signifies that both parties want to share their public key
)

type Message struct {
	Protocol int

	Session_ID string
	Data []byte

	PB_Key []byte
}
