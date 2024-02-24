package torrent

import (
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type Torrent struct {
	Announce    string
	InfoHash    []byte
	Pieces      [][]byte
	PieceLength int
	Length      int
	Name        string
}

type Peer struct {
	IP   net.IP
	Port uint16
}

// Parses a torrent (AKA metainfo) file into a structure
// The format of a torrent file can be found here:
// https://www.bittorrent.org/beps/bep_0003.html#metainfo-files
func ParseTorrent(filename string) (Torrent, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return Torrent{}, err
	}

	res, _, err := DecodeBencode(string(bytes))
	if err != nil {
		return Torrent{}, err
	}

	// Assume torrent file contains correct keys
	metainfo := res.(map[string]interface{})
	info := metainfo["info"].(map[string]interface{})
	bencodedInfo := metainfo["info bencoded"].(string)

	// Calculate SHA-1 hash of the bencoded info dictionary
	hasher := sha1.New()
	hasher.Write([]byte(bencodedInfo))

	// Populate torrent structure
	var file Torrent
	file.Announce = metainfo["announce"].(string)
	file.InfoHash = hasher.Sum(nil)
	file.Pieces = splitPieces(info["pieces"].(string), 20)
	file.PieceLength = info["piece length"].(int)
	file.Length = info["length"].(int)
	file.Name = info["name"].(string)
	return file, nil
}

// Helper function to get SHA-1 piece hashes
func splitPieces(pieces string, chunkLength int) [][]byte {
	if len(pieces)%chunkLength != 0 {
		panic("invalid pieces")
	}
	var chunk []byte
	chunks := make([][]byte, 0, len(pieces)/chunkLength)
	for len(pieces) >= chunkLength {
		chunk, pieces = []byte(pieces[:chunkLength]), pieces[chunkLength:]
		chunks = append(chunks, chunk)
	}
	return chunks
}

// Gets the peers of a torrent by sending a GET request to the torrent tracker
func (torrent Torrent) GetPeers(peerId []byte) (string, error) {
	base, err := url.Parse(torrent.Announce)
	if err != nil {
		return "", err
	}
	query := url.Values{
		"info_hash":  []string{string(torrent.InfoHash)},
		"peer_id":    []string{string(peerId[:])},            // Assume length of 20 bytes
		"port":       []string{"6881"},                       // Default port for downloading
		"uploaded":   []string{"0"},                          // Assume zero for now
		"downloaded": []string{"0"},                          // Assume zero for now
		"left":       []string{strconv.Itoa(torrent.Length)}, // Assume full length for now
		"compact":    []string{"1"},
	}
	base.RawQuery = query.Encode()

	res, err := http.Get(base.String())
	if err != nil {
		return "", err
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "", err
	}
	if res.StatusCode != 200 {
		return "", errors.New("request to peer failed with: " + res.Status)
	}

	return string(body), nil
}

// Parses the peers of a torrent given the correct bencode
func ParsePeers(bencode string) ([]Peer, error) {
	res, _, err := DecodeBencode(bencode)
	if err != nil {
		return []Peer{}, nil
	}

	// Ignore interval key for now
	info := res.(map[string]interface{})
	strPeers := info["peers"].(string)

	const peerSize int = 6
	const ipSize int = 4
	if len(strPeers)%peerSize != 0 {
		panic("invalid peers")
	}
	peers := make([]Peer, 0, len(strPeers)/peerSize)
	var strPeer []byte
	var peer Peer
	for len(strPeers) >= peerSize {
		strPeer, strPeers = []byte(strPeers[:peerSize]), strPeers[peerSize:]
		peer.IP = net.IP(strPeer[:ipSize])
		peer.Port = binary.BigEndian.Uint16(strPeer[ipSize:])
		peers = append(peers, peer)
	}
	return peers, nil
}
