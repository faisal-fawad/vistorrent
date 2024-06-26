package torrent

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"
)

type Work struct {
	Index  int
	Length int
}

type Result struct {
	Index  int
	Result []byte
}

func DownloadFile(name string, destination string, w http.ResponseWriter) error {
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
	workQueue := make(chan *Work, len(torr.PieceHashes))
	resQueue := make(chan *Result)
	for i := range torr.PieceHashes {
		size := int(torr.PieceLength)
		// Get size and reduce if necessary (can occur on last piece of data)
		if int(torr.Length)-i*size < size {
			size = int(torr.Length) - i*size
		}
		workQueue <- &Work{i, size}
	}

	for i := range peers {
		var peer Peer = peers[i]
		go torr.PieceWorker(peer, peerId, workQueue, resQueue)
	}

	// Send number of pieces to server
	fmt.Fprintf(w, "data: %d \n\n", len(torr.PieceHashes))
	w.(http.Flusher).Flush()
	time.Sleep(1 * time.Second)
	// For case study

	done := 0
	total := len(torr.PieceHashes)
	file := make([]byte, torr.Length)
	for done < total {
		res := <-resQueue
		bound := int(torr.PieceLength) * res.Index
		copy(file[bound:bound+len(res.Result)], res.Result)
		done++
		fmt.Printf("Piece #%d complete (%d / %d) with %d peers \n", res.Index, done, total, runtime.NumGoroutine()-1)

		// Send data to server
		fmt.Fprintf(w, "data: %d \n\n", res.Index)
		w.(http.Flusher).Flush()
		// For case study
	}
	close(workQueue)
	close(resQueue)

	err = os.WriteFile(destination, file, 0644)
	return err
}
