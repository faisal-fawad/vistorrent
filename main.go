package main

import (
	"fmt"
	"os"

	"github.com/faisal-fawad/vistorrent/decode"
)

func main() {
	torrent, err := decode.ParseTorrent(os.Args[1])
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
}
