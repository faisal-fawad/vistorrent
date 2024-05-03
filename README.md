# vistorrent
A lightweight torrenting client built in Go with the main goal of gaining a deeper understanding of torrenting! **Not recommended for 🌊🚢🏴‍☠️**

## Getting Started

### Dependencies
- Git
- Golang

### Installation & Execution
- Clone or download *this* repository
- Build the project by running `go build`
- Run the project with `./vistorrent <input:file> <output:file>`
- Navigate to `http://localhost:8080` and start the download by clicking the button

## Demo
https://github.com/faisal-fawad/vistorrent/assets/76597599/4dfd4308-f9f8-4aa3-a5d3-f9ec20f48d6c

Each red box represents a piece of a file, when that piece has been downloaded, it turns green! If a peer fails to download a piece, it is placed back onto the work queue (hence the appearance of "missed" red boxes in the demo)

## Future Plans
- Support for magnet links (currently only supports `.torrent` files)
- Support for other tracker types and/or a [distributed hash table](https://www.bittorrent.org/beps/bep_0005.html) (currently only supports HTTP trackers)
- Support for multi-file torrents (currently only supports single-file torrents)
- Support for seeding (currently only supports leeching)
- Make visualization optional and use a desktop application instead of a web application
