package torrent

import (
	"crypto/rand"
	"fmt"
	"os"
)

type work struct {
	Index  int
	Length int
}

type result struct {
	Index  int
	Result []byte
}

func DownloadFile(name string, destination string) error {
	torr, err := ParseTorrent(name)
	if err != nil {
		return err
	}

	// Get peers by using a random peerId
	peerId := make([]byte, peerIdSize)
	rand.Read(peerId)
	peers, err := torr.GetPeers(peerId)
	if err != nil {
		return err
	}

	// Make channels for each piece
	workQueue := make(chan *work, len(torr.PieceHashes))
	resQueue := make(chan *result, len(torr.PieceHashes))
	for i := range torr.PieceHashes {
		size := int(torr.PieceLength)
		// Get size and reduce if necessary (can occur on last piece of data)
		if int(torr.Length)-(i*size) < size {
			size = int(torr.Length) - (i * size)
		}
		workQueue <- &work{i, size}
	}

	for i := range peers {
		var peer Peer = peers[i]
		go torr.PieceWorker(peer, peerId, workQueue, resQueue)
	}

	done := 0
	total := len(torr.PieceHashes)
	file := make([]byte, torr.Length)
	for done < total {
		res := <-resQueue
		bound := int(torr.PieceLength) * res.Index
		copy(file[bound:bound+len(res.Result)], res.Result)
		fmt.Printf("Piece %d: %x \n", res.Index, res.Result)
		done++
	}
	close(workQueue)
	close(resQueue)

	fmt.Printf("Code reaching here means success (hopefully)!")
	err = os.WriteFile(destination, file, 0644)
	return err
}
