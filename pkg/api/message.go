package api

const (
	INITIAL_CONNECT = 0 // Signifies the message is for sending a passphrase to relay server for PAKE
	CONFIRMATION = 1 // Signifies that both clients (A and B) have connected to the relay server so that the PAKE procss can begin
	SEND_FILE = 2 // Signifies the message is for sending data to end user
)

type Message struct {
	Protocol int

	Session_ID string
	Data       []byte
}
