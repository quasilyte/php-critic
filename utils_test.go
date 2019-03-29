package main

import (
	"testing"
)

func TestInterpretString(t *testing.T) {
	type testCase struct {
		quote byte
		raw   string
		want  string
	}
	type simpleTestCase struct {
		raw  string
		want string
	}

	sharedTests := []simpleTestCase{
		{"", ""},
		{"a", "a"},
		{"abc", "abc"},

		{`\\`, `\`},
		{`\\\\`, `\\`},
		{`\\x\\x`, `\x\x`},
	}

	// Strings enclosed between ''.
	q1tests := []simpleTestCase{
		{`\$x`, `\$x`},
		{`\'`, `'`},
		{`\"`, `\"`},
		{`\n`, `\n`},
		{`\r\n`, `\r\n`},
		{`\x00`, `\x00`},
		{`\xff`, `\xff`},
		{`\x1aaa`, `\x1aaa`},
	}

	// Strings enclosed between "".
	q2tests := []simpleTestCase{
		{`\$x`, `$x`},
		{`\'`, `\'`},
		{`\"`, `"`},
		{`\n`, "\n"},
		{`\r\n`, "\r\n"},
		{`\x00`, "\x00"},
		{`\xff`, "\xff"},
		{`\x1aaa`, "\x1aaa"},
	}

	var tests []testCase
	for _, test := range sharedTests {
		tests = append(tests,
			testCase{quote: '"', raw: test.raw, want: test.want},
			testCase{quote: '\'', raw: test.raw, want: test.want})
	}
	for _, test := range q1tests {
		tests = append(tests, testCase{quote: '\'', raw: test.raw, want: test.want})
	}
	for _, test := range q2tests {
		tests = append(tests, testCase{quote: '"', raw: test.raw, want: test.want})
	}

	for _, test := range tests {
		want := test.want
		have, ok := interpretString(test.raw, test.quote)
		if !ok {
			t.Errorf("evalString(%q, %v): failed to eval",
				test.raw, test.quote)
			continue
		}
		if have != want {
			t.Errorf("evalString(%q, %v): results mismatch:\nhave: %q\nwant: %q",
				test.raw, test.quote, have, want)
		}
	}
}
