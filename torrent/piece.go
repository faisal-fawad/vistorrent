package torrent

type Message struct {
	Type    MessageType
	Payload []byte
}

type MessageType int

const (
	Choke         MessageType = 0 // No payload
	Unchoke       MessageType = 1 // No payload
	Interested    MessageType = 2 // No payload
	NotInterested MessageType = 3 // No payload
	Have          MessageType = 4 // Payload contains index
	Bitfield      MessageType = 5 // Payload consists of a singular byte
	Request       MessageType = 6 // Payload contains index, begin, and length
	Piece         MessageType = 7 // Same payload as request
	Cancel        MessageType = 8 // Payload contains index, begin, and piece
)

func (t *Torrent) DownloadPiece(peer Peer, peerId []byte) {
	// Add code here!
}
