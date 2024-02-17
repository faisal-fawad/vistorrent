package main

import (
	"fmt"
	"os"

	"github.com/faisal-fawad/vistorrent/parse"
)

func main() {
	torrent, err := parse.ParseTorrent(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	// Print structure
	fmt.Println("Announce: " + torrent.Announce)
	fmt.Printf("InfoHash: %x \n", torrent.InfoHash)
	fmt.Printf("PieceLength: %d \n", torrent.PieceLength)
	fmt.Printf("Length: %d \n", torrent.Length)
	fmt.Println("Name: " + torrent.Name)
	fmt.Println("Pieces: ")
	for i := range torrent.Pieces {
		fmt.Printf("%x \n", torrent.Pieces[i])
	}

	var peerId [20]byte
	copy(peerId[:], "00112233445566778899")
	fmt.Println("Peers Bencode: ")
	bencode, err := torrent.GetPeers(peerId)
	if err != nil {
		fmt.Println(err)
		return
	}
	peers, err := parse.ParsePeers(bencode)
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := range peers {
		fmt.Printf("%s:%d \n", peers[i].IP, peers[i].Port)
	}
}
