package ksyn

import (
	"fmt"
	"testing"
)

func TestParseInt(t *testing.T) {
	tests := []struct {
		Input  string
		Output int
	}{
		{"1", 1},
		{"100", 100},
		{"50", 50},
		{"123", 123},
		{"0123", 123},
		{"1230", 1230},
		{"666 ", 666},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("subtest%d", i), func(t *testing.T) {
			p := parser{
				source: tt.Input,
			}
			out, err := p.num()
			if err != nil {
				t.Error(err)
			}
			i, ok := out.(Int)
			if !ok {
				t.Error(fmt.Errorf("Expected Int: Got %T", out))
				return
			}
			if i.Int != tt.Output {
				t.Error(fmt.Errorf("Expected integer %#v: Got %#v", tt.Output, i.Int))
				return
			}
			return
		})
	}

	return
}
