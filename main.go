package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/faisal-fawad/vistorrent/parse"
)

func main() {
	var filename string = "debian.iso.torrent"
	bytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	decoded, _, err := parse.DecodeBencode(string(bytes))
	if err != nil {
		fmt.Println(err)
		return
	}
	out, _ := json.Marshal(decoded)
	fmt.Println(string(out))
}
