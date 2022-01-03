package ksyn

import (
	"fmt"
)

type Dict struct {
	Entries    []DictEntry
	start, end Pos
}

type DictEntry struct {
	Key, Val   Expr
	start, end Pos
}

func (p *parser) exprDict() (Dict, error) {
	start := p.pos

	if err := p.consumeRune('{'); err != nil {
		return Dict{}, err
	}

	r := Dict{
		start: start,
	}

	for {
		p.chomp()
		if p.ch() == '}' {
			p.next()
			break
		}

		entrystart := p.pos

		k, err := p.expr()
		if err != nil {
			return Dict{}, err
		}

		p.chomp()

		if err := p.consumeRune(':'); err != nil {
			return Dict{}, err
		}

		p.chomp()

		v, err := p.expr()
		if err != nil {
			return Dict{}, err
		}

		entryend := p.pos

		r.Entries = append(r.Entries, DictEntry{
			Key:   k,
			Val:   v,
			start: entrystart,
			end:   entryend,
		})

		//p.chomp()
		//if p.ch() == '}' {
		//	continue
		//}
		//if err := p.consumeRune(','); err != nil {
		//	return Dict{}, err
		//}
	}

	return r, nil
}

func (d Dict) Start() Pos {
	return d.start
}
func (d Dict) End() Pos {
	return d.end
}

func (de DictEntry) Start() Pos {
	return de.start
}
func (de DictEntry) End() Pos {
	return de.end
}

func (d Dict) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1
	r += fmt.Sprintf("%T\n", d)

	for i, de := range d.Entries {
		if i+1 == len(d.Entries) {
			r += de.repr(style, pre2+style.Leaf, pre2+style.Blank)
		} else {
			r += de.repr(style, pre2+style.Branch, pre2+style.Carry)
		}
	}

	return r
}

func (de DictEntry) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1
	r += fmt.Sprintf("%T\n", de)

	r += de.Key.repr(style, pre2+style.Branch, pre2+style.Carry)
	r += de.Val.repr(style, pre2+style.Leaf, pre2+style.Blank)

	return r
}
