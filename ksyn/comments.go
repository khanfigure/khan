package ksyn

import (
	"fmt"
	"strings"
)

type CommentStyle int

const (
	InvalidCommentStyle CommentStyle = iota
	SingleLineComment
	MultiLineComment
)

func (cs CommentStyle) String() string {
	switch cs {
	case SingleLineComment:
		return "SingleLineComment"
	case MultiLineComment:
		return "MultiLineComment"
	default:
		return "InvalidCommentStyle"
	}
}

type Comment struct {
	Comment string
	Style   CommentStyle

	start, end Pos
}

func (c Comment) Start() Pos {
	return c.start
}
func (c Comment) End() Pos {
	return c.end
}

func (p *parser) comments() ([]Comment, error) {
	var r []Comment
	for {
		v, err := p.comment()
		if err != nil {
			return nil, err
		}
		// no (more) comments found
		if v.Style == InvalidCommentStyle {
			return r, nil
		}

		// SUPER NOT SURE ABOUT THIS
		p.chomp()

		r = append(r, v)
	}
}

func (p *parser) comment() (Comment, error) {
	str := p.strPeek()
	if strings.HasPrefix(str, "//") {
		p.space()

		start := p.pos

		if err := p.consume("//"); err != nil {
			return Comment{}, err
		}

		str = p.str()
		for i := 0; i < len(str); i++ {
			if str[i] == '\n' {
				c := Comment{
					Comment: str[:i],
					Style:   SingleLineComment,
					start:   start,
					end:     p.pos,
				}
				if err := p.consumeRune('\n'); err != nil {
					return Comment{}, err
				}
				return c, nil
			}
			p.next()
		}
		return Comment{}, p.unexpectedRune("new line to finish single line comment")
	} else if strings.HasPrefix(str, "/*") {
		p.space()

		start := p.pos

		if err := p.consume("/*"); err != nil {
			return Comment{}, err
		}

		str = p.str()
		for i := 0; i < len(str); i++ {
			if strings.HasPrefix(p.str(), "*/") {
				c := Comment{
					Comment: str[:i],
					Style:   MultiLineComment,
					start:   start,
					end:     p.pos,
				}
				if err := p.consume("*/"); err != nil {
					return Comment{}, err
				}
				return c, nil
			}
			p.next()
		}
		return Comment{}, p.unexpectedRune("*/ to finish multi-line comment")
	}

	return Comment{}, nil
}

func (p *parser) strPeek() string {
	str := p.str()
	for i := 0; i < len(str); i++ {
		if str[i] != ' ' && str[i] != '\t' {
			return str[i:]
		}
	}
	return ""
}
func (p *parser) strPeekChomp() string {
	str := p.str()
	for i := 0; i < len(str); i++ {
		if str[i] != ' ' && str[i] != '\t' && str[i] != '\n' {
			return str[i:]
		}
	}
	return ""
}

func (p *parser) strPeekChompComments() string {
	str := p.str()
	commentMode := 0
	for i := 0; i < len(str); i++ {
		switch commentMode {
		case 0:
			if strings.HasPrefix(str[i:], "//") {
				commentMode = 1
				i++
			} else if strings.HasPrefix(str[i:], "/*") {
				commentMode = 2
				i++
			} else {
				if str[i] != ' ' && str[i] != '\t' && str[i] != '\n' {
					return str[i:]
				}
			}
		case 1:
			if str[i] == '\n' {
				commentMode = 0
			}
		case 2:
			if strings.HasPrefix(str[i:], "*/") {
				commentMode = 0
				i++
			}
		}
	}
	return ""
}

func (c Comment) repr(style reprStyle, pre1, pre2 string) string {
	return pre1 + fmt.Sprintf("%s %#v\n", c.Style, c.Comment)
}
