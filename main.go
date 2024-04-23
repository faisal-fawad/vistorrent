package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/faisal-fawad/vistorrent/torrent"
)

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	// Set headers (may want to change in a production environment)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Start downloading file
	torrent.DownloadFile(os.Args[1], os.Args[2], w)
	os.Exit(0)
}

func main() {
	if len(os.Args[1:]) != 2 {
		fmt.Println("invoke this command by using: ./vistorrent <input:file> <output:file>")
		return
	}
	fmt.Println("serving on http://localhost:8080")
	http.Handle("/", http.FileServer(http.Dir("frontend")))
	http.HandleFunc("/download", downloadHandler)
	http.ListenAndServe(":8080", nil)
}
