package ksyn

import ()

type reprStyle struct {
	Blank  string
	Branch string
	Carry  string
	Leaf   string

	Comments bool
}

var unicodeReprStyle = reprStyle{
	Blank:  "   ",
	Branch: " ├─ ",
	Carry:  " │  ",
	Leaf:   " └─ ",

	Comments: true,
}

var asciiReprStyle = reprStyle{
	Blank:  "    ",
	Branch: " +- ",
	Carry:  " |  ",
	Leaf:   " '- ",

	Comments: true,
}

var minimalReprStyle = reprStyle{
	Blank:  "  ",
	Branch: "  ",
	Carry:  "  ",
	Leaf:   "  ",

	Comments: true,
}

func ReprUnicode(n Node) string {
	return n.repr(unicodeReprStyle, "", "")
}
func ReprAscii(n Node) string {
	return n.repr(asciiReprStyle, "", "")
}
func ReprMinimal(n Node) string {
	return n.repr(minimalReprStyle, "", "")
}

func repr_title(title string, node Node, style reprStyle, pre1, pre2 string) string {
	r := pre1 + title + "\n"
	r += node.repr(style, pre2+style.Leaf, pre2+style.Blank)
	return r
}

/*
	case If:
		r += "If\n"
		for _, pair := range v.Pairs {
			r += repr(pair.Expr, indent)
			r += repr(pair.Block, indent)
		}
		r += repr(v.Else, indent)
	case Expr:
		r += "Expr\n"
		r += repr(v.PreComments, indent)
		r += repr(v.Expr, indent)
		r += repr(v.PostComments, indent)
	case Op:
		r += "Op(" + v.Operator.String() + ")\n"
		r += repr(v.Left, indent)
		r += repr(v.PreComments, indent)
		r += repr(v.PostComments, indent)
		r += repr(v.Right, indent)
	case Map:
		r += "Map\n"
		r += repr(v.Expr, indent)
		r += repr(v.Block, indent)
	case Ident:
		r += fmt.Sprintf("Ident(%s)\n", v.Ident)
	case *Ident:
		if v == nil {
			r += "*Ident (nil)\n"
		} else {
			r += "*Ident\n"
			r += repr(*v, indent)
		}
	case Int:
		r += fmt.Sprintf("Int(%d)\n", v.Int)
	case Str:
		r += fmt.Sprintf("Str(%#v)\n", v.Str)
	case RGB:
		r += fmt.Sprintf("RGB(#%02x%02x%02x)\n", v.R, v.G, v.B)
	case Call:
		r += "Call\n"
		r += repr(v.Ident, indent)
		r += repr(v.Args, indent)
	case Func:
		r += "Func\n"
		r += repr(v.Ident, indent)
		r += repr(v.Args, indent)
		if len(v.Returns) > 0 {
			r += ir + ii + "Returns\n"
			r += repr(v.Returns, indent+1)
		}
		r += repr(v.Block, indent)
	case []Var:
		if len(v) == 0 {
			return ""
		}
		r += "Vars\n"
		for _, vv := range v {
			r += repr(vv, indent)
		}
	case Var:
		r += "Var\n"
		r += repr(v.Ident, indent)
		r += repr(v.Type, indent)
	case Block:
		r += "Block\n"
		r += repr(v.PreComments, indent)
		for _, stmt := range v.Stmts {
			r += repr(stmt, indent)
		}
		r += repr(v.PostComments, indent)
	case *Block:
		if v == nil {
			r += "*Block (nil)\n"
		} else {
			r += "*Block\n"
			r += repr(*v, indent)
		}
	case []Comment:
		if len(v) == 0 {
			return ""
		}
		r += "Comments\n"
		for _, comment := range v {
			r += repr(comment, indent)
		}
	case Comment:
		r += fmt.Sprintf("%s(%#v)\n", v.Style, v.Comment)
	case []Node:
		if len(v) == 0 {
			return ""
		}
		r += "Nodes\n"
		for _, node := range v {
			r += repr(node, indent)
		}
	case Select:
		r += "Select\n"
		for _, vv := range v.Exprs {
			r += repr(vv, indent)
		}
		r += repr(v.From, indent)
	case SelectExpr:
		r += "SelectExpr\n"
		if v.Table.Ident != "" {
			r += repr(v.Table, indent)
		}
		r += repr(v.Column, indent)
		if v.As.Ident != "" {
			r += repr(v.As, indent)
		}
	case SelectFrom:
		r += "SelectFrom\n"
		r += repr(v.Table, indent)
		if v.As.Ident != "" {
			r += repr(v.As, indent)
		}
	default:
		r += fmt.Sprintf("unhandled(%T)\n", n)
	}
	return r
}*/
