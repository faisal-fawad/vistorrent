package torrent

import (
	"encoding/binary"
	"fmt"
)

const lengthSize int = 4

type Message struct {
	Length  uint32
	Type    byte
	Payload []byte
}

const (
	Choke         byte = 0 // No payload
	Unchoke       byte = 1 // No payload
	Interested    byte = 2 // No payload
	NotInterested byte = 3 // No payload
	Have          byte = 4 // Payload contains index
	Bitfield      byte = 5 // Payload consists of a singular byte
	Request       byte = 6 // Payload contains index, begin, and length
	Piece         byte = 7 // Same payload as request
	Cancel        byte = 8 // Payload contains index, begin, and piece
)

// Builds a []byte representation of message
func (m *Message) BuildMessage() []byte {
	res := make([]byte, 4, m.Length+5)
	binary.BigEndian.PutUint32(res, m.Length)
	res = append(res, m.Type)
	res = append(res, m.Payload...)
	return res
}

// Parses a message into a structure
func ParseMessage(stream []byte) (Message, error) {
	var m Message
	length := binary.BigEndian.Uint32(stream[:lengthSize])
	if int(length)+4 != len(stream) {
		return Message{}, fmt.Errorf("error parsing message: expected: %d -> got: %d", int(length)+4, len(stream))
	}

	m.Length = length
	m.Type = byte(stream[lengthSize])
	m.Payload = stream[lengthSize+1:]
	return m, nil
}

// Handles all message types in the message structure
func (m *Message) HandleMessage() {
	switch m.Type {
	case Choke:
		fmt.Println("Choked!")
	case Unchoke:
		fmt.Println("Unchoked!")
	case Interested:
		fmt.Println("Interested!")
	case NotInterested:
		fmt.Println("Not interested!")
	case Have:
		fmt.Println("Have!")
	case Bitfield:
		fmt.Printf("Bitfield! %b \n", m.Payload)
	case Request:
		fmt.Println("Request!")
	case Piece:
		fmt.Println("Piece!")
	case Cancel:
		fmt.Println("Cancel!")
	}
}

func (t *Torrent) DownloadPiece(peer Peer, peerId []byte) error {
	// Do handshake
	conn, _, err := peer.PeerHandshake(t.InfoHash, peerId)
	if err != nil {
		return err
	}
	defer conn.Close()

	conn.Write([]byte{Unchoke})
	conn.Write([]byte{Interested})

	buf, err := ReadFullWithLength(conn, 4, 0)
	if err != nil {
		return err
	}
	msg, err := ParseMessage(buf)
	if err != nil {
		return err
	}

	fmt.Printf("Message (bytes): %x \n", msg.BuildMessage())
	msg.HandleMessage()
	return nil
}
