package ksyn

import (
	"fmt"
	"testing"
)

func TestStrPeekChomp(t *testing.T) {
	tests := []struct {
		Input  string
		Output string
	}{
		{"", ""},
		{" ", ""},
		{"a ", "a "},
		{" a", "a"},
		{"a", "a"},
		{"   abc", "abc"},
		{"abc   ", "abc   "},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("subtest%d", i), func(t *testing.T) {
			p := parser{
				source: tt.Input,
			}
			out := p.strPeekChomp()
			if out != tt.Output {
				t.Error(fmt.Errorf("Expected strchomp() output %#v: Got %#v", tt.Output, out))
			}
			return
		})
	}

	return
}
