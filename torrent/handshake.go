package torrent

import (
	"fmt"
	"net"
	"strconv"
)

type Handshake struct {
	PeerId   []byte
	InfoHash []byte
}

func (peer Peer) PeerHandshake() (Handshake, error) {
	conn, err := net.Dial("tcp", peer.IP.String()+":"+strconv.Itoa(int(peer.Port)))
	if err != nil {
		return Handshake{}, err
	}
	var stream []byte
	bytes, err := conn.Read(stream)
	if err != nil {
		return Handshake{}, err
	}
	fmt.Printf("%x \n", bytes)
	return Handshake{}, nil
}
