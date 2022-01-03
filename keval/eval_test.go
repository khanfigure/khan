package keval

import (
	"fmt"
	"strings"
	"testing"

	"button/lex"
	"button/parse"
)

func TestEval(t *testing.T) {
	type testcase struct {
		title string
		in    string
		out   string
		err   string
	}
	testcases := []testcase{
		{"EmptyInput", "", "", ""},
		{"IntegerLiteral", "666", "666", ""},
		{"Addition", "1 + 2", "3", ""},
		{"Subtraction", "7 - 10", "-3", ""},
		{"OperatorPrecedence", "666 + 666 * 666", "444222", ""},
		{"Parenthesis", "(666 + 666) * 666", "887112", ""},
		{"BooleanNegation", "!true", "false", ""},
		{"Scope1", `
b := true
a := func() {
	if b {
		"b is true"
	} else {
		"b is false"
	}
}
a()
`, `b is true`, ``},
		{"Scope2", `
a := func() {
	if b {
		"b is true"
	} else {
		"b is false"
	}
}
b := true
a()
`, ``, `unknown identifier "b"`},
		{"Scope3", `
a := 1
{
	a := 2
	{
		a := a + 1
		{
			a
		}
	}
}`, `3`, ``},
	}
	for ti, tc := range testcases {
		title := tc.title
		if title == "" {
			title = fmt.Sprintf("test_%d", ti)
		}
		t.Run(title, func(t *testing.T) {
			out, err := func() (string, error) {
				items := lex.Lex(tc.in, title)
				node, err := parse.Parse(items)
				if err != nil {
					return "", err
				}
				val, err := Eval(node)
				if err != nil {
					return "", err
				}
				if val.Type == parse.ValNone {
					return "", nil
				}
				if val.Type == parse.ValString {
					return val.Str, nil
				}
				return val.String(), nil
			}()
			errstr := ""
			if err != nil {
				errstr = err.Error()
			}
			if !strings.Contains(strings.ToLower(errstr), strings.ToLower(tc.err)) {
				t.Error(fmt.Errorf("Expected error substring %#v: Got %#v", tc.err, errstr))
				return
			}
			if out != tc.out {
				t.Error(fmt.Errorf("Expected output %#v: Got %#v", tc.out, out))
				return
			}
		})
	}

}
