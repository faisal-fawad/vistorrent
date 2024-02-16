package decode

import (
	"strconv"
	"strings"
	"unicode"
)

const msg string = " this string may not follow the bencode schema"

// Decodes a bencode string into its respective type: string, int, slice, or map
// NOTE: this function does not handle bounds out of range errors
func DecodeBencode(bencode string) (interface{}, int, error) {
	if unicode.IsDigit(rune(bencode[0])) {
		// Parse string -> string
		i := strings.Index(bencode, ":")
		if i < 0 {
			panic("':' not found!" + msg)
		}

		length, err := strconv.Atoi(bencode[:i])
		if err != nil {
			return "", 0, err
		}

		return bencode[i+1 : i+1+length], i + 1 + length, nil
	} else if bencode[0] == 'i' {
		// Parse integer -> int
		i := strings.Index(bencode, "e")
		if i < 0 {
			panic("'e' not found!" + msg)
		}

		res, err := strconv.Atoi(bencode[1:i])
		if err != nil {
			return "", 0, err
		}

		return res, i + 1, nil
	} else if bencode[0] == 'l' {
		// Parse list -> slice
		var i int = 1
		slice := make([]interface{}, 0, 5)

		for bencode[i] != 'e' {
			res, len, err := DecodeBencode(bencode[i:])
			if err != nil {
				return "", 0, err
			}

			slice = append(slice, res)
			i += len
		}

		return slice, i + 1, nil
	} else if bencode[0] == 'd' {
		// Parse dictionary -> map
		var i int = 1
		dict := make(map[string]interface{})

		for bencode[i] != 'e' {
			key, keyLen, err := DecodeBencode(bencode[i:])
			if err != nil {
				return "", 0, err
			}
			keyStr, ok := key.(string)
			if !ok {
				panic("key not string!" + msg)
			}
			i += keyLen

			val, valLen, err := DecodeBencode(bencode[i:])
			if err != nil {
				return "", 0, err
			}
			i += valLen

			dict[keyStr] = val
		}

		return dict, i + 1, nil
	} else {
		panic("'%q' unexpected!" + msg)
	}
}
