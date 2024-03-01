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
	fmt.Println("Pieces: ")
	for i := range torr.Pieces {
		fmt.Printf("%x \n", torr.Pieces[i])
	}

	fmt.Println("Peers Bencode: ")
	peers, err := torr.GetPeers([]byte("00112233445566778899"))
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := range peers {
		fmt.Printf("%s:%d \n", peers[i].IP, peers[i].Port)
		peers[i].PeerHandshake(torr.InfoHash, []byte("00112233445566778899"))
	}
}
