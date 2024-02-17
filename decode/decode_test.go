package decode

import (
	"reflect"
	"testing"
)

type decodeTest struct {
	bencode  string
	expected interface{}
}

var allTests = []decodeTest{
	// Strings
	{"11:hello world", "hello world"},
	{"4:spam", "spam"},
	{"3:egg", "egg"},
	{"100:", ""},
	// Integers
	{"i3e", 3},
	{"i-3e", -3},
	{"i100e", 100},
	// Lists (composition of above)
	{"l4:spam4:eggse", []interface{}{"spam", "eggs"}},
	{"l11:hello worldi3ei-3ee", []interface{}{"hello world", 3, -3}},
	// Dictionaries (composition of above)
	{"d3:cow3:moo4:spam4:eggse", map[string]interface{}{"cow": "moo", "spam": "eggs"}},
	{"d4:spaml1:a1:bee", map[string]interface{}{"spam": []interface{}{"a", "b"}}},
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
