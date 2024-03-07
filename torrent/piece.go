package torrent

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

const lengthSize int = 4
const requestLength int = 12 // 3 32-bit integers
const blockSize uint32 = 16384

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
	Bitfield      byte = 5 // Payload consists of bytes (bits) for piece possession
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
// This is just to help with readability, may be removed in the future
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
		fmt.Println("Bitfield!")
	case Request:
		fmt.Println("Request!")
	case Piece:
		fmt.Println("Piece!")
	case Cancel:
		fmt.Println("Cancel!")
	default:
		fmt.Println("Invalid message!")
	}
}

const pieceIndex uint32 = 0 // Currently constant, will change once implemented fully

// Downloads a piece by communicating with the peer
// All integers sent through the BitTorrent protocol at this point
// are encoded as 4 bytes big endian
func (t *Torrent) DownloadPiece(peer Peer, peerId []byte) ([]byte, error) {
	// Do handshake
	conn, _, err := peer.PeerHandshake(t.InfoHash, peerId)
	if err != nil {
		return []byte{}, err
	}
	defer conn.Close()

	// Read the bitfield and -> TODO: build a simple data structure to store which pieces a peer has available
	buf, err := ReadFullWithLength(conn, 4, 0)
	if err != nil {
		return []byte{}, err
	}
	msg, err := ParseMessage(buf)
	if err != nil {
		return []byte{}, err
	}
	msg.HandleMessage()

	// Write interested and unchoked, since connections start choked and uninterested
	interested := Message{1, Interested, nil}
	unchoke := Message{1, Unchoke, nil}
	conn.Write(interested.BuildMessage())
	conn.Write(unchoke.BuildMessage())

	// The peer should send back an unchoke message, which we will wait for with a blocking call
	buf, err = ReadFullWithLength(conn, 4, 0) // Blocking call
	if err != nil {
		return []byte{}, err
	}
	msg, err = ParseMessage(buf)
	if err != nil {
		return []byte{}, err
	}
	fmt.Printf("Message (bytes): %x \n", buf)
	msg.HandleMessage()

	// Download the block
	piece, err := t.DownloadBlock(conn)
	if err != nil {
		return []byte{}, err
	}

	// Check validity of piece
	if !t.ValidPiece(piece, int(pieceIndex)) {
		return []byte{}, errors.New("piece invalid")
	}

	return piece, nil
}

// Helper function to download the blocks of a piece
// We are assuming that when entering this function the peer is ready to
// communicate via request and piece messages. TODO: implement pipelining
func (t *Torrent) DownloadBlock(conn net.Conn) ([]byte, error) {
	// Break the piece into blocks of 16 KiB (2^14 = 16384 bytes)
	var i uint32
	piece := make([]byte, t.PieceLength)
	for i = 0; i*blockSize < t.PieceLength; i++ {
		requestPayload := BuildRequestPayload(i, t.PieceLength)

		// Write request
		request := Message{uint32(requestLength + 1), Request, requestPayload}
		conn.Write(request.BuildMessage())

		// Read piece
		buf, err := ReadFullWithLength(conn, 4, 0)
		if err != nil {
			return []byte{}, err
		}
		msg, err := ParseMessage(buf)
		if err != nil {
			return []byte{}, err
		}
		_, offset, block := ParsePiecePayload(msg.Payload)

		// Place the block into its corresponding location
		copy(piece[offset:offset+blockSize], block)
	}
	return piece, nil
}

// Helper function to handle request payloads
func BuildRequestPayload(i uint32, pieceLength uint32) []byte {
	requestPayload := make([]byte, requestLength)

	// Get size and reduce if necessary (can occur on last block of data)
	var size uint32 = blockSize
	if pieceLength-i*blockSize < blockSize {
		size = pieceLength - i*blockSize
	}

	// Populate payload
	binary.BigEndian.PutUint32(requestPayload, pieceIndex)
	binary.BigEndian.PutUint32(requestPayload[4:], i*blockSize)
	binary.BigEndian.PutUint32(requestPayload[8:], size)

	return requestPayload
}

// Helper function to handle piece payloads
func ParsePiecePayload(payload []byte) (uint32, uint32, []byte) {
	// Parse data from payload
	index := binary.BigEndian.Uint32(payload[:4])
	begin := binary.BigEndian.Uint32(payload[4:8])
	var block []byte = payload[8:]

	return index, begin, block
}

func (t *Torrent) ValidPiece(piece []byte, index int) bool {
	hash := GetHash(piece)
	return bytes.Equal(hash, t.PieceHashes[index])
}
