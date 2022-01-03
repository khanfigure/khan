package ksyn

import (
	"fmt"
)

type Select struct {
	Exprs []SelectExpr
	From  SelectFrom

	start, end Pos
}

func (s Select) Start() Pos {
	return s.start
}
func (s Select) End() Pos {
	return s.end
}

type SelectExpr struct {
	Table  Ident
	Column Ident
	As     Ident

	start, end Pos
}

func (se SelectExpr) Start() Pos {
	return se.start
}
func (se SelectExpr) End() Pos {
	return se.end
}
func (se SelectExpr) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1
	r += fmt.Sprintf("%T\n", se)

	r += se.Table.repr(style, pre2+style.Branch, pre2+style.Carry)
	r += se.Column.repr(style, pre2+style.Branch, pre2+style.Carry)
	r += se.As.repr(style, pre2+style.Leaf, pre2+style.Blank)

	return r
}

type SelectFrom struct {
	Table Ident
	As    Ident

	start, end Pos
}

func (sf SelectFrom) Start() Pos {
	return sf.start
}
func (sf SelectFrom) End() Pos {
	return sf.end
}
func (sf SelectFrom) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1
	r += fmt.Sprintf("%T\n", sf)

	r += sf.Table.repr(style, pre2+style.Branch, pre2+style.Carry)
	r += sf.As.repr(style, pre2+style.Leaf, pre2+style.Blank)

	return r
}

func (p *parser) parseSelect() (Select, error) {
	start := p.pos

	if err := p.consumeFold("select"); err != nil {
		return Select{}, err
	}

	var exprs []SelectExpr

	for {
		expr, err := p.parseSelectExpr()
		if err != nil {
			return Select{}, err
		}

		exprs = append(exprs, expr)

		p.chomp()

		if p.ch() == ',' {
			if err := p.consume(","); err != nil {
				return Select{}, err
			}
		} else {
			break
		}
	}

	p.chomp()

	if err := p.consumeFold("from"); err != nil {
		return Select{}, err
	}

	p.chomp()

	from, err := p.parseSelectFrom()
	if err != nil {
		return Select{}, err
	}

	end := p.pos
	return Select{
		Exprs: exprs,
		From:  from,
		start: start,
		end:   end,
	}, nil
}

func (p *parser) parseSelectExpr() (SelectExpr, error) {
	p.chomp()

	start := p.pos

	ident, err := p.ident()
	if err != nil {
		return SelectExpr{}, err
	}

	var (
		col Ident
		as  Ident
	)

	if p.ch() == '.' {
		if err := p.consume("."); err != nil {
			return SelectExpr{}, err
		}

		col, err = p.ident()
		if err != nil {
			return SelectExpr{}, err
		}
	}

	p.chomp()

	if hasPrefixFold(p.str(), "as") {
		if err := p.consumeFold("as"); err != nil {
			return SelectExpr{}, err
		}
		p.chomp()

		as, err = p.ident()
		if err != nil {
			return SelectExpr{}, err
		}
	}

	end := p.pos

	se := SelectExpr{
		start: start,
		end:   end,
	}

	if col.Ident != "" {
		se.Table = ident
		se.Column = col
	}
	se.As = as

	return se, nil
}

func (p *parser) parseSelectFrom() (SelectFrom, error) {
	p.chomp()

	start := p.pos

	ident, err := p.ident()
	if err != nil {
		return SelectFrom{}, err
	}

	p.chomp()

	var as Ident

	if hasPrefixFold(p.str(), "as") {
		if err := p.consumeFold("as"); err != nil {
			return SelectFrom{}, err
		}
		p.chomp()

		as, err = p.ident()
		if err != nil {
			return SelectFrom{}, err
		}
	}

	end := p.pos
	return SelectFrom{
		Table: ident,
		As:    as,

		start: start,
		end:   end,
	}, nil
}

func (s Select) repr(style reprStyle, pre1, pre2 string) string {
	r := pre1
	r += fmt.Sprintf("%T\n", s)

	for _, sf := range s.Exprs {
		r += sf.Table.repr(style, pre2+style.Branch, pre2+style.Carry)
	}
	r += s.From.repr(style, pre2+style.Leaf, pre2+style.Blank)

	return r
}
