package torrent

import (
	"crypto/sha1"
	"os"
)

const hashLength int = 20

// A torrent structure, note that the sizes of InfoHash and PieceHashes are not explictly
// defined to make working with them easier. Typically, a hash has a length of 20, which is
// constantly defined above.
type Torrent struct {
	Announce    string
	InfoHash    []byte
	PieceHashes [][]byte
	PieceLength uint32 // Can't be negative
	Length      uint32 // Also can't be negative
	Name        string
}

// A structure to define errors that occur with parsing a torrent file
type TorrentError struct {
	err string
}

func (t *TorrentError) Error() string {
	return t.err
}

// Parses a torrent (AKA metainfo) file into a structure
// The format of a torrent file can be found here:
// https://www.bittorrent.org/beps/bep_0003.html#metainfo-files
func ParseTorrent(filename string) (Torrent, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return Torrent{}, &TorrentError{"failed to read file"}
	}

	res, _, err := DecodeBencode(string(bytes))
	if err != nil {
		return Torrent{}, &TorrentError{"invalid bencode: " + err.Error()}
	}

	// Populate the torrent structure using type assertion
	metainfo := res.(map[string]interface{})
	info := metainfo["info"].(map[string]interface{})
	if metainfo == nil || info == nil {
		return Torrent{}, &TorrentError{"bencode missing keys"}
	}

	var file Torrent
	file.Announce = metainfo["announce"].(string)
	strInfoHash := metainfo["info bencoded"].(string)
	strPieces := info["pieces"].(string)
	file.PieceLength = uint32(info["piece length"].(int))
	file.Length = uint32(info["length"].(int))
	file.Name = info["name"].(string)
	if file.Announce == "" || strInfoHash == "" || strPieces == "" || file.PieceLength == 0 || file.Length == 0 || file.Name == "" {
		return Torrent{}, &TorrentError{"bencode missing values"}
	}

	// Calculate SHA-1 hash of the bencoded info dictionary and split piece hashes
	file.InfoHash = GetHash([]byte(strInfoHash))
	file.PieceHashes, err = SplitPieces(strPieces, hashLength)
	if err != nil {
		return Torrent{}, &TorrentError{err.Error()}
	}

	return file, nil
}

// Helper function to split a string on every multiple of n (chunkLength)
func SplitPieces(pieces string, chunkLength int) ([][]byte, error) {
	if len(pieces)%chunkLength != 0 {
		return [][]byte{}, &TorrentError{"invalid pieces"}
	}

	var chunk []byte
	chunks := make([][]byte, 0, len(pieces)/chunkLength)
	for len(pieces) >= chunkLength {
		chunk, pieces = []byte(pieces[:chunkLength]), pieces[chunkLength:]
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// Helper function to calculate the SHA-1 hash
func GetHash(data []byte) []byte {
	hasher := sha1.New()
	hasher.Write([]byte(data))
	return hasher.Sum(nil)
}
