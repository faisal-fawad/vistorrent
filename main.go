package main

import (
	"fmt"
	"os"

	"github.com/faisal-fawad/vistorrent/torrent"
)

func main() {
	torr, err := torrent.ParseTorrent(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	// Print structure
	fmt.Println("Announce: " + torr.Announce)
	fmt.Printf("InfoHash: %x \n", torr.InfoHash)
	fmt.Printf("PieceLength: %d \n", torr.PieceLength)
	fmt.Printf("Length: %d \n", torr.Length)
	fmt.Println("Name: " + torr.Name)
	peers, err := torr.GetPeers([]byte("00112233445566778899"))
	if err != nil {
		fmt.Println(err)
		return
	}

	// For now, lets work with the first peer only
	err = torr.DownloadPiece(peers[0], []byte("00112233445566778899"))
	if err != nil {
		fmt.Println(err)
	}
}
