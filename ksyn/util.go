package ksyn

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/mgutz/ansi"
)

// Error is an error attached to a source position
type Error struct {
	Err error
	Pos Pos
}

func (ee Error) Unwrap() error {
	return ee.Err
}
func (ee Error) Error() string {
	if ee.Pos.Path != "" {
		return fmt.Sprintf("%s %s", ee.Pos, ee.Err.Error())
	}
	return ee.Err.Error()
}

// ErrorFromNode is an error attached to an emitted node
type ErrorFromNode struct {
	Err  error
	Node Node
}

func (efn ErrorFromNode) Unwrap() error {
	return efn.Err
}
func (efn ErrorFromNode) Error() string {
	return efn.Err.Error()
}
func NewErrFromNode(node Node, err error) error {
	return ErrorFromNode{
		Err:  err,
		Node: node,
	}
}
func Errorf(node Node, errf string, args ...interface{}) error {
	return ErrorFromNode{
		Err:  fmt.Errorf(errf, args...),
		Node: node,
	}
}

func (p parser) unexpectedRune(expected string) error {
	str := p.str()
	for _, r := range str {
		if r > 32 && r < 127 {
			return Error{
				Pos: p.pos,
				Err: fmt.Errorf("Unexpected %s: Expected %s", string(r), expected),
			}
		}
		return Error{
			Pos: p.pos,
			Err: fmt.Errorf("Unexpected rune %#v (0x%x): Expected %s", string(r), r, expected),
		}
	}
	return Error{
		Pos: p.pos,
		Err: fmt.Errorf("Unexpected EOF: Expected %s", expected),
	}
}

func (p parser) unexpected(expected string) error {
	ch := p.ch()
	if ch == 0 {
		return Error{
			Pos: p.pos,
			Err: fmt.Errorf("Unexpected EOF: Expected %s", expected),
		}
	}
	//panic("here")
	return Error{
		Pos: p.pos,
		Err: fmt.Errorf("Unexpected syntax: Expected %s", expected),
	}
}

func (p parser) ch() byte {
	if p.pos.Offset >= len(p.source) {
		return 0
	}
	return p.source[p.pos.Offset]
}

func (p parser) str() string {
	if p.pos.Offset >= len(p.source) {
		return ""
	}
	return p.source[p.pos.Offset:]
}

// next is like move(1) but correctly accounts for newlines increasing the
// line and column number.
func (p *parser) next() {
	pos := p.pos
	ch := p.source[pos.Offset]
	pos.Offset++
	pos.Col++
	if ch == '\n' {
		pos.Line++
		pos.Col = 0
	}
	p.pos = pos
}

// don't move through newlines.  move does not do as much processing
// as next() does.
func (p *parser) move(delta int) {
	pos := p.pos
	pos.Offset += delta
	pos.Col += delta
	p.pos = pos
}

// don't call consume() with an argument that has newlines in it
func (p *parser) consume(s string) error {
	if !strings.HasPrefix(p.str(), s) {
		return p.unexpected(fmt.Sprintf("%#v", s))
	}
	p.move(len(s))
	return nil
}

func (p *parser) consumeFold(s string) error {
	if !hasPrefixFold(p.str(), s) {
		return p.unexpected(fmt.Sprintf("%#v", s))
	}
	p.move(len(s))
	return nil
}

func (p *parser) consumeRune(r byte) error {
	ch := p.ch()
	if ch != r {
		return p.unexpectedRune(string(r))
	}
	p.next()
	return nil
}

func (p *parser) space() {
	for {
		ch := p.ch()
		if ch == ' ' || ch == '\t' {
			p.move(1)
			continue
		}
		break
	}
}

func (p *parser) chomp() {
	for {
		ch := p.ch()
		if ch == ' ' || ch == '\t' || ch == '\n' {
			p.next()
			continue
		}
		break
	}
}

func hasPrefixToken(in string, search string) bool {
	if in == search {
		return true
	}
	if !strings.HasPrefix(in, search) {
		return false
	}
	if ch_is_ident_rest(in[len(search)]) {
		return false
	}
	return true
}

func NodeSourceFragment(n Node) string {
	start := n.Start()
	end := n.End()
	return SourceFragmentRange(start, end)
}

func SourceFragment(pos Pos) string {
	end := pos
	return SourceFragmentRange(pos, end)
}

func width(in string) int {
	w := runewidth.StringWidth(in)
	for _, r := range in {
		if r == '\t' {
			w += 3
		}
	}
	return w
}
func tabfix(in string) string {
	return strings.ReplaceAll(in, "\t", "   ")
}

func seekLineStart(c Pos, buf []byte) Pos {
	for c.Offset > 0 {
		if buf[c.Offset-1] == '\n' {
			break
		}
		c.Offset--
	}
	c.Col = 0
	return c
}
func seekPreviousLine(c Pos, buf []byte) Pos {
	// this function should only be called after a call to
	// seekLineStart, otherwise line/col counts will be off.
	if c.Offset > 0 && buf[c.Offset-1] == '\n' {
		c.Offset--
		c.Line--
		c = seekLineStart(c, buf)
	}
	return c
}
func seekNextLine(c Pos, buf []byte) Pos {
	for c.Offset+1 < len(buf) {
		c.Offset++
		c.Col++
		if buf[c.Offset-1] == '\n' {
			c.Col = 0
			c.Line++
			break
		}
	}
	return c
}
func seekNextChar(c Pos, buf []byte) Pos {
	if c.Offset+1 < len(buf) {
		if buf[c.Offset] == '\n' {
			c.Col = 0
			c.Line++
		} else {
			c.Col++
		}
		c.Offset++
	}
	return c
}

func SourceFragmentRange(start, end Pos) string {
	if start.Path == "" {
		return ""
	}

	buf, err := ioutil.ReadFile(start.Path)
	if err != nil || len(buf) == 0 {
		return ""
	}

	long := true

	if start.Offset == end.Offset {
		end = seekNextChar(end, buf)
		long = false
	}

	pre_lines := 3
	post_lines := 3

	showstart := seekLineStart(start, buf)
	for i := 0; i < pre_lines; i++ {
		showstart = seekPreviousLine(showstart, buf)
	}

	showend := seekNextLine(end, buf)
	for i := 0; i < post_lines; i++ {
		showend = seekNextLine(showend, buf)
	}

	r := ""

	red := ansi.ColorCode("black:red")
	reset := ansi.ColorCode("reset")

	//	marker := ""
	//	markerSize := width(string(buf[start.Offset:end.Offset]))
	//	for i := 0; i < markerSize; i++ {
	//		marker += "▲"
	//	}

	c := showstart
	for c.Offset < showend.Offset {
		e := seekNextLine(c, buf)
		if e.Offset == c.Offset {
			break
		}

		s := string(buf[c.Offset:e.Offset])

		s = strings.TrimSuffix(s, "\r")
		s = strings.TrimSuffix(s, "\n")

		si := 0

		if long {
			if c.Line == start.Line {
				s = s[0:start.Col] + red + s[start.Col:]
				si += len(red)
			}
			if c.Line == end.Line {
				s = s[0:end.Col+si] + reset + s[end.Col+si:]
			}
		}

		//		if c.Line == start.Line + 1 {
		//s = s[0:start.Col] + red(marker) + s[start.Col + markerSize - 1:len(s)]
		//		}

		r += fmt.Sprintf("%8d │ %s\n", c.Line+1, s)
		c = e
	}

	/*	header := fmt.Sprintf("%8d │ ", start.Line + 1)
		pad := ""
		for i := 0; i < width(header); i++ {
			pad += " "
		}
		for i := 0; i < width(string(buf[linestart:start.Offset])); i++ {
			pad += " "
		}
		marker := ""
		markerSize := width(string(buf[start.Offset:end.Offset]))
		for i := 0; i < markerSize; i++ {
			marker += "▲"
		}

		return header + tabfix(string(buf[linestart:start.Offset])) +
			tabfix(string(buf[start.Offset:end.Offset])) +
			tabfix(string(buf[end.Offset:lineend])) + "\n" + pad + marker*/

	return r
}
