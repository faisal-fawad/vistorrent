package torrent

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

const extensionSize int = 8
const peerIdSize int = 20

type Handshake struct {
	ProtocolLength byte
	Protocol       string
	Extensions     []byte
	PeerId         []byte
	InfoHash       []byte
}

// Builds a []byte representation of handshake
func (h *Handshake) BuildHandshake() []byte {
	res := make([]byte, 0, len(h.Protocol)+len(h.Extensions)+len(h.PeerId)+len(h.InfoHash)+1)
	res = append(res, h.ProtocolLength)
	res = append(res, []byte(h.Protocol)...)
	res = append(res, h.Extensions...)
	res = append(res, h.PeerId...)
	res = append(res, h.InfoHash...)
	return res
}

// Parses a handshake into a structure
func ParseHandshake(stream []byte) (Handshake, error) {
	var h Handshake
	length := stream[0]
	if len(stream) != int(length)+extensionSize+hashLength+peerIdSize+1 {
		return Handshake{}, fmt.Errorf("error parsing handshake: expected: %d -> got: %d", int(length)+extensionSize+hashLength+peerIdSize+1, len(stream))
	}

	h.ProtocolLength = length
	h.Protocol = string(stream[1 : length+1])
	h.Extensions = stream[length+1 : int(length)+extensionSize+1]
	h.PeerId = stream[int(length)+extensionSize+1 : int(length)+extensionSize+peerIdSize+1]
	h.InfoHash = stream[int(length)+extensionSize+peerIdSize+1:]

	return h, nil
}

func (peer Peer) PeerHandshake(infoHash []byte, peerId []byte) (net.Conn, Handshake, error) {
	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
	if err != nil {
		return nil, Handshake{}, err
	}

	// Send handshake
	var inHand Handshake = Handshake{
		byte(19),
		"BitTorrent protocol",
		make([]byte, 8),
		infoHash,
		peerId,
	}
	var in []byte = inHand.BuildHandshake()
	_, err = conn.Write(in)
	if err != nil {
		return nil, Handshake{}, err
	}
	fmt.Printf("In <- %x \n", in)

	// Receive handshake
	out, err := ReadFullWithLength(conn, 1, hashLength+peerIdSize+extensionSize)
	if err != nil {
		return nil, Handshake{}, err
	}

	outHand, err := ParseHandshake(out)
	if err != nil {
		return nil, Handshake{}, err
	}
	fmt.Printf("Out -> %x \n", outHand.BuildHandshake())

	return conn, outHand, nil
}

// Helper function to read data from a TCP connection where the data uses the following schema:
// A number of bytes n to indicate the size of the message, which is then followed by n bytes of data
// This function also provides a parameter for extra bytes to support dealing with certain types of messages
// An example message looks like: 0x0000000205e0 -> 4 bytes to represent the length of the payload (0x00000002),
// which in decimal representation is 2, followed by 2 bytes of data (0x05e0)
func ReadFullWithLength(conn net.Conn, prefixLength int, extraBytes int) ([]byte, error) {
	if prefixLength > 4 || prefixLength < 0 {
		return []byte{}, errors.New("prefix length must be in the range 0-4")
	}
	bufLength := make([]byte, prefixLength)
	_, err := io.ReadFull(conn, bufLength)
	if err != nil {
		return []byte{}, err
	}

	lengthSlice := make([]byte, 4-prefixLength, 4)
	lengthSlice = append(lengthSlice, bufLength...)
	var length int = int(binary.BigEndian.Uint32(lengthSlice))
	if err != nil {
		return []byte{}, err
	}
	buf := make([]byte, length+extraBytes)
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		return []byte{}, err
	}

	return append(bufLength, buf...), nil
}
