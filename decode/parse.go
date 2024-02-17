package decode

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
)

type Torrent struct {
	Announce    string ``
	InfoHash    []byte
	Pieces      [][]byte
	PieceLength int
	Length      int
	Name        string
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

	jsonBytes, _ := json.Marshal(res)
	fmt.Println(string(jsonBytes))

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
	var chunk []byte
	chunks := make([][]byte, 0, len(pieces)/chunkLength)
	for len(pieces) >= chunkLength {
		chunk, pieces = []byte(pieces[:chunkLength]), pieces[chunkLength:]
		chunks = append(chunks, chunk)
	}
	if len(pieces) > 0 {
		chunks = append(chunks, []byte(pieces))
	}
	return chunks
}
