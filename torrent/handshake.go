package torrent

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

type Handshake struct {
	ProtocolLength int
	Protocol       string
	Extensions     []byte
	PeerId         []byte
	InfoHash       []byte
}

func (peer Peer) PeerHandshake(infoHash []byte, peerId []byte) (Handshake, error) {
	conn, err := net.DialTimeout("tcp", peer.IP.String()+":"+strconv.Itoa(int(peer.Port)), 3*time.Second)
	if err != nil {
		return Handshake{}, err
	}
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.Close()

	var in []byte = []byte("\x13BitTorrent protocol\x00\x00\x00\x00\x00\x00\x00\x00" + string(infoHash) + string(peerId))
	_, err = conn.Write(in)
	if err != nil {
		return Handshake{}, err
	}
	fmt.Printf("In <- %x \n", in)

	var out []byte = make([]byte, 1024)
	n, err := conn.Read(out)
	if err != nil {
		return Handshake{}, err
	}
	fmt.Printf("Out -> %x \n", out[:n])
	return Handshake{}, nil
}
