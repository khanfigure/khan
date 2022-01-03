package ksyn

import (
	"strings"
)

func hasPrefixFold(in string, search string) bool {
	return len(in) >= len(search) && strings.EqualFold(in[:len(search)], search)
}
