package torrent

import (
	"errors"
	"fmt"
	"net"
	"time"
)

const extensionSize = 8
const peerIdSize = 20
const bufSize = 1024

type Handshake struct {
	ProtocolLength byte
	Protocol       string
	Extensions     []byte
	PeerId         []byte
	InfoHash       []byte
}

// Builds the default handshake
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
	var length byte = stream[0]
	if len(stream) != int(length)+extensionSize+peerIdSize+hashLength+1 {
		return Handshake{}, errors.New("error parsing handshake")
	}

	h.ProtocolLength = length
	h.Protocol = string(stream[1 : length+1])
	h.Extensions = stream[length+1 : length+extensionSize+1]
	h.PeerId = stream[length+extensionSize+1 : length+extensionSize+peerIdSize+1]
	h.InfoHash = stream[length+extensionSize+peerIdSize+1:]

	return h, nil
}

func (peer Peer) PeerHandshake(infoHash []byte, peerId []byte) (net.Conn, Handshake, error) {
	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
	if err != nil {
		return nil, Handshake{}, err
	}
	conn.SetDeadline(time.Now().Add(3 * time.Second))

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

	var out []byte = make([]byte, bufSize)
	n, err := conn.Read(out)
	if err != nil {
		return nil, Handshake{}, err
	}
	outHand, err := ParseHandshake(out[:n])
	if err != nil {
		return nil, Handshake{}, err
	}
	fmt.Printf("Out -> %x \n", outHand.BuildHandshake())

	return conn, outHand, nil
}
