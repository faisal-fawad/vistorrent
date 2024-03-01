package torrent

import (
	"crypto/sha1"
	"errors"
	"os"
)

const hashLength int = 20

type Torrent struct {
	Announce    string
	InfoHash    []byte
	Pieces      [][]byte
	PieceLength int
	Length      int
	Name        string
}

type TorrentError struct {
	Err error
}

func (t *TorrentError) Error() string {
	return t.Err.Error()
}

// Parses a torrent (AKA metainfo) file into a structure
// The format of a torrent file can be found here:
// https://www.bittorrent.org/beps/bep_0003.html#metainfo-files
func ParseTorrent(filename string) (Torrent, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return Torrent{}, &TorrentError{err}
	}

	res, _, err := DecodeBencode(string(bytes))
	if err != nil {
		return Torrent{}, &TorrentError{err}
	}

	// Populate the torrent structure using type assertion
	metainfo := res.(map[string]interface{})
	info := metainfo["info"].(map[string]interface{})
	if metainfo == nil || info == nil {
		return Torrent{}, &TorrentError{errors.New("key not found")}
	}

	var file Torrent
	file.Announce = metainfo["announce"].(string)
	strInfoHash := metainfo["info bencoded"].(string)
	strPieces := info["pieces"].(string)
	file.PieceLength = info["piece length"].(int)
	file.Length = info["length"].(int)
	file.Name = info["name"].(string)
	if file.Announce == "" || strInfoHash == "" || strPieces == "" || file.PieceLength == 0 || file.Length == 0 || file.Name == "" {
		return Torrent{}, &TorrentError{errors.New("value not found")}
	}

	// Calculate SHA-1 hash of the bencoded info dictionary and split piece hashes
	hasher := sha1.New()
	hasher.Write([]byte(strInfoHash))
	file.InfoHash = hasher.Sum(nil)
	file.Pieces, err = SplitPieces(strPieces, hashLength)
	if err != nil {
		return Torrent{}, &TorrentError{err}
	}

	return file, nil
}

// Helper function to get SHA-1 piece hashes
func SplitPieces(pieces string, chunkLength int) ([][]byte, error) {
	if len(pieces)%chunkLength != 0 {
		return [][]byte{}, &TorrentError{errors.New("invalid pieces")}
	}

	var chunk []byte
	chunks := make([][]byte, 0, len(pieces)/chunkLength)
	for len(pieces) >= chunkLength {
		chunk, pieces = []byte(pieces[:chunkLength]), pieces[chunkLength:]
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}
