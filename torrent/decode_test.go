package torrent

import (
	"reflect"
	"testing"
)

type decodeTest struct {
	bencode  string
	expected interface{}
}

// An empty string means that we are expecting some sort of error
var allTests = []decodeTest{
	// Strings
	{"11:hello world", "hello world"},
	{"4:spam", "spam"},
	{"3:egg", "egg"},
	{"100", ""},
	{"1abc:", ""},
	{"100:", ""},
	// Integers
	{"i3e", 3},
	{"i-3e", -3},
	{"i100e", 100},
	{"i100", ""},
	{"iabce", ""},
	// Lists (composition of above)
	{"l4:spam4:eggse", []interface{}{"spam", "eggs"}},
	{"l11:hello worldi3ei-3ee", []interface{}{"hello world", 3, -3}},
	{"le", []interface{}{}},
	{"l", ""},
	{"li100", ""},
	{"li100e", ""},
	// Dictionaries (composition of above)
	{"d3:cow3:moo4:spam4:eggse", map[string]interface{}{"cow": "moo", "spam": "eggs"}},
	{"d4:spaml1:a1:bee", map[string]interface{}{"spam": []interface{}{"a", "b"}}},
	{"d4:info8:bencodede", map[string]interface{}{"info": "bencoded", "info bencoded": "8:bencoded"}},
	{"de", map[string]interface{}{}},
	{"d", ""},
	{"dabc", ""},
	{"di100e", ""},
	{"d3:cow", ""},
	{"d3:cowabc", ""},
	{"d3:cow3:moo", ""},
}

func TestDecodeBencode(t *testing.T) {
	for _, test := range allTests {
		got, _, _ := DecodeBencode(test.bencode)
		var equal bool
		switch got.(type) {
		case string, int:
			equal = got == test.expected
		case []interface{}, map[string]interface{}:
			equal = reflect.DeepEqual(got, test.expected)
		default:
			equal = false
		}
		if !equal {
			t.Errorf("expected: %q -> got: %q", test.expected, got)
		}
	}
}
