package torrent

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const msg string = "! this string may not follow the bencode schema"
const start int = 1

// A structure to define errors that occur with decoding bencode
type DecodeError struct {
	err string
}

func (d *DecodeError) Error() string {
	return d.err + msg
}

// Decodes a bencode string into its respective type in Go: string, int, slice, or map
// The bencode schema (similar to JSON) can be found here:
// https://www.bittorrent.org/beps/bep_0003.html#bencoding
func DecodeBencode(bencode string) (interface{}, int, error) {
	if unicode.IsDigit(rune(bencode[0])) {
		// Parse string -> string
		i := strings.Index(bencode, ":")
		if i < 0 {
			return "", 0, &DecodeError{"':' not found"}
		}

		length, err := strconv.Atoi(bencode[:i])
		if err != nil {
			return "", 0, &DecodeError{"int not found"}
		}

		if i+1+length > len(bencode) {
			return "", 0, &DecodeError{"index out of bounds"}
		}

		return bencode[i+1 : i+1+length], i + 1 + length, nil
	} else if bencode[0] == 'i' {
		// Parse integer -> int
		i := strings.Index(bencode, "e")
		if i < 0 {
			return "", 0, &DecodeError{"'e' not found"}
		}

		res, err := strconv.Atoi(bencode[1:i])
		if err != nil {
			return "", 0, &DecodeError{"int not found"}
		}

		return res, i + 1, nil
	} else if bencode[0] == 'l' {
		// Parse list -> slice
		if start > len(bencode)-1 {
			return "", 0, &DecodeError{"index out of bounds"}
		}
		var i int = start
		slice := make([]interface{}, 0, 5)

		for bencode[i] != 'e' {
			res, length, err := DecodeBencode(bencode[i:])
			if err != nil {
				return "", 0, err
			}

			slice = append(slice, res)
			i += length

			if i > len(bencode)-1 {
				return "", 0, &DecodeError{"index out of bounds"}
			}
		}

		return slice, i + 1, nil
	} else if bencode[0] == 'd' {
		// Parse dictionary -> map
		if start > len(bencode)-1 {
			return "", 0, &DecodeError{"index out of bounds"}
		}
		var i int = start
		dict := make(map[string]interface{})

		for bencode[i] != 'e' {
			// Key
			key, keyLen, err := DecodeBencode(bencode[i:])
			if err != nil {
				return "", 0, err
			}
			keyStr, ok := key.(string)
			if !ok {
				return "", 0, &DecodeError{"key not string"}
			}

			i += keyLen
			if i > len(bencode)-1 {
				return "", 0, &DecodeError{"index out of bounds"}
			}

			// Value
			val, valLen, err := DecodeBencode(bencode[i:])
			if err != nil {
				return "", 0, err
			}
			dict[keyStr] = val
			// This is needed for info hash verification later in the BitTorrent protocol
			if keyStr == "info" {
				dict[keyStr+" bencoded"] = bencode[i : i+valLen]
			}

			i += valLen
			if i > len(bencode)-1 {
				return "", 0, &DecodeError{"index out of bounds"}
			}
		}

		return dict, i + 1, nil
	} else {
		return "", 0, &DecodeError{fmt.Sprintf("'%q' unexpected", bencode[0])}
	}
}
