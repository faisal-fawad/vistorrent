package torrent

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"
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

// Downloads a piece by communicating with a specified peer.
// All integers sent through the BitTorrent protocol are encoded as 4 bytes big endian
func (t *Torrent) PieceWorker(peer Peer, peerId []byte, workQueue chan *work, resQueue chan *result) error {
	// Do handshake
	conn, _, err := peer.PeerHandshake(t.InfoHash, peerId)
	if err != nil {
		return err
	}
	conn.SetDeadline(time.Now().Add(time.Second * 10))
	defer conn.Close()

	// Read the bitfield to later check which pieces have the piece we're after
	buf, err := ReadFullWithLength(conn, 4, 0)
	if err != nil {
		return err
	}
	msg, err := ParseMessage(buf)
	if err != nil {
		return err
	}
	var bitfield []byte = msg.Payload

	// Write interested and unchoked, since connections start choked and uninterested
	interested := Message{1, Interested, nil}
	unchoke := Message{1, Unchoke, nil}
	conn.Write(interested.BuildMessage())
	conn.Write(unchoke.BuildMessage())

	// The peer should send back an unchoke message here, which we will then read
	buf, err = ReadFullWithLength(conn, 4, 0)
	if err != nil {
		return err
	}
	msg, err = ParseMessage(buf)
	if err != nil {
		return err
	}
	msg.HandleMessage()

	// Attempt to download a piece from the work queue
	for work := range workQueue {
		if !HavePiece(bitfield, work.Index) {
			workQueue <- work // Place work back on queue
			continue
		}

		piece, err := t.DownloadBlock(conn, work)
		if err != nil {
			workQueue <- work // Place work back on queue
			fmt.Println(err)
			return err
		}

		if !t.ValidatePiece(piece, work.Index) {
			workQueue <- work // Place work back on queue
			fmt.Println("Failed integrity check!")
			continue
		}

		// Piece is complete and valid, place on to the results queue
		resQueue <- &result{work.Index, piece}
	}
	return nil
}

// Helper function to download the blocks of a piece
// We are assuming that when entering this function the peer is ready to
// communicate via request and piece messages. TODO: implement pipelining
func (t *Torrent) DownloadBlock(conn net.Conn, work *work) ([]byte, error) {
	// Break the piece into blocks of 16 KiB (2^14 = 16384 bytes)
	var i uint32
	piece := make([]byte, work.Length)
	for i = 0; i*blockSize < t.PieceLength; i++ {
		requestPayload := BuildRequestPayload(i, uint32(work.Length), uint32(work.Index))

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
		copy(piece[offset:int(offset)+len(block)], block)
	}
	return piece, nil
}

// Helper function to handle request payloads
func BuildRequestPayload(i uint32, pieceLength uint32, pieceIndex uint32) []byte {
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

// Helper function to check if a hash of a piece matches with its corresponding
// hash from the torrent structure
func (t *Torrent) ValidatePiece(piece []byte, index int) bool {
	hash := GetHash(piece)
	return bytes.Equal(hash, t.PieceHashes[index])
}

// Helper function to check if a peers bitfield contains a specific piece
func HavePiece(bitfield []byte, index int) bool {
	byteIndex := index / 8
	byteOffset := index % 8
	return bitfield[byteIndex]>>(7-byteOffset)&1 != 0
}
