package main

import (
	"fmt"
	"os"

	"github.com/faisal-fawad/vistorrent/torrent"
)

func main() {
	err := torrent.DownloadFile(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Println(err)
	}
}
