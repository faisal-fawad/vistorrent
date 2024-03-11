package torrent

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
)

const peerSize int = 6
const ipSize int = peerSize - 2

type Peer struct {
	IP   net.IP
	Port uint16
}

// Converts the peer to its string representation
func (peer *Peer) String() string {
	return fmt.Sprintf("%s:%d", peer.IP.String(), peer.Port)
}

// A structure to define errors that occur with networking
type NetworkError struct {
	err string
}

func (n *NetworkError) Error() string {
	return n.err
}

// Gets the peers of a torrent by sending a GET request to the torrent tracker
func (torrent *Torrent) GetPeers(peerId []byte) ([]Peer, error) {
	// Build the url
	base, err := url.Parse(torrent.Announce)
	if err != nil {
		return []Peer{}, &DecodeError{"unable to parse tracker URL"}
	}
	query := url.Values{
		"info_hash":  []string{string(torrent.InfoHash)},
		"peer_id":    []string{string(peerId)},             // Assume length of 20 bytes
		"port":       []string{"6881"},                     // Default port for downloading
		"uploaded":   []string{"0"},                        // Assume zero for now
		"downloaded": []string{"0"},                        // Assume zero for now
		"left":       []string{fmt.Sprint(torrent.Length)}, // Assume full length for now
		"compact":    []string{"1"},
	}
	base.RawQuery = query.Encode()

	// Send GET request
	res, err := http.Get(base.String())
	if err != nil {
		return []Peer{}, &NetworkError{"failed to get peers"}
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body) // On success, body contains bencode
	if err != nil {
		return []Peer{}, err
	}
	if res.StatusCode != http.StatusOK {
		return []Peer{}, &NetworkError{fmt.Sprintf("failed to get peers with status: %s\nand body: %s", res.Status, body)}
	}

	peers, err := ParsePeers(string(body))
	if err != nil {
		return []Peer{}, &DecodeError{err.Error()}
	}

	return peers, nil
}

// Helper function that parses peers into a structure
func ParsePeers(bencode string) ([]Peer, error) {
	res, _, err := DecodeBencode(bencode)
	if err != nil {
		return []Peer{}, err
	}

	// TODO: implement interval key, ignored for now
	info := res.(map[string]interface{})
	strPeers := info["peers"].(string)

	// Populate torrent structure array
	peers := make([]Peer, 0, len(strPeers)/peerSize)
	var bytePeers [][]byte
	bytePeers, err = SplitPieces(strPeers, peerSize)
	if err != nil {
		return []Peer{}, err
	}

	for i := range bytePeers {
		var peer Peer
		peer.IP = net.IP(bytePeers[i][:ipSize])
		peer.Port = binary.BigEndian.Uint16(bytePeers[i][ipSize:])
		peers = append(peers, peer)
	}

	return peers, nil
}
