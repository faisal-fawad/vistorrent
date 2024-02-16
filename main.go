package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/faisal-fawad/vistorrent/decode"
)

func main() {
	var filename string = "debian.iso.torrent"
	bytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	decoded, _, err := decode.DecodeBencode(string(bytes))
	if err != nil {
		fmt.Println(err)
		return
	}
	out, _ := json.Marshal(decoded)
	fmt.Println(string(out))
}
