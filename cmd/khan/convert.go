package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"

	"khan.rip"

	"gopkg.in/yaml.v3"
)

const (
	khanpkgname  = "khan.rip"
	khanpkgalias = "khan"
)

type yamlwalker struct {
	br *buildrun

	gobuf   *string
	assetfs *bool

	imports  map[string]string
	yamlpath string
	wd       string

	shortkey   string
	shortvalue string
}

type yamlerror struct {
	path string
	node *yaml.Node
	err  error
}

func (err yamlerror) Error() string {
	return fmt.Sprintf("%s:%d:%d: %v", err.path, err.node.Line, err.node.Column, err.err)
}

type yamlhandler func(w *yamlwalker, v *yaml.Node) error

var yamlhandlers = map[string]yamlhandler{
	"file":  yamlsimplehandler(&khan.File{}),
	"group": yamlsimplehandler(&khan.Group{}),
	"user":  yamlsimplehandler(&khan.User{}),
}

func yamlkind(kind yaml.Kind) string {
	switch kind {
	case yaml.DocumentNode:
		return "document"
	case yaml.SequenceNode:
		return "array"
	case yaml.MappingNode:
		return "map"
	case yaml.ScalarNode:
		return "scalar"
	case yaml.AliasNode:
		return "alias"
	default:
		return fmt.Sprintf("unknown yaml.Kind %#v", kind)
	}
}

func (w *yamlwalker) nodeErrorf(node *yaml.Node, format string, a ...interface{}) error {
	return yamlerror{
		path: w.yamlpath,
		node: node,
		err:  fmt.Errorf(format, a...),
	}
}

func (w *yamlwalker) addimport(pkg, alias string) string {
	suffix := 0
	newalias := alias
	for {
		clear := true
		for k, v := range w.imports {
			if k == pkg {
				return v
			}
			if v == newalias {
				suffix++
				newalias = fmt.Sprintf("%s%d", alias, suffix)
				clear = false
				break
			}
		}
		if clear {
			w.imports[pkg] = newalias
			return newalias
		}
	}
}

func (w *yamlwalker) yamlwalk(node *yaml.Node) error {
	if node.Kind == 0 {
		// A document with nothing but comments seems to return this
		return nil
	}

	if node.Kind != yaml.DocumentNode {
		return w.nodeErrorf(node, "Expected document: Got %s", yamlkind(node.Kind))
	}

	for _, child := range node.Content {
		if err := w.yamlwalkdoc(child); err != nil {
			return err
		}
	}

	return nil
}

func (w *yamlwalker) yamlwalkdoc(node *yaml.Node) error {

	if node.Kind == yaml.ScalarNode && node.Value == "" {
		// A document with nothing but a --- header seems to return this
		return nil
	}

	if node.Kind == yaml.SequenceNode {
		for _, child := range node.Content {
			if err := w.yamlwalkdoc(child); err != nil {
				return err
			}
		}
		return nil
	}
	if node.Kind == yaml.MappingNode {
		if len(node.Content)%2 != 0 {
			return w.nodeErrorf(node, "Odd sized YAML map")
		}
		for i := 0; i < len(node.Content); i += 2 {
			k := node.Content[i]
			v := node.Content[i+1]
			if k.Kind != yaml.ScalarNode {
				return w.nodeErrorf(k, "Expected scalar map key: Got %s", yamlkind(k.Kind))
			}

			// TODO maybe pass these to handlerfunc to have better scoping
			w.shortkey = ""
			w.shortvalue = ""

			handler := k.Value

			spc := strings.IndexByte(handler, ' ')
			if spc != -1 {
				w.shortkey = handler[spc:]
				handler = handler[:spc-1]
			}

			// special super-shortcut for files
			if w.shortkey == "" && strings.HasPrefix(handler, "/") {
				w.shortkey = handler
				handler = "file"
			}

			h, ok := yamlhandlers[handler]
			if !ok {
				return w.nodeErrorf(k, "Invalid khan-yaml type %#v", handler)
			}

			if err := h(w, v); err != nil {
				return err
			}
		}
		return nil
	}
	return w.nodeErrorf(node, "Expected array or map: Got %s", yamlkind(node.Kind))
}

func (br *buildrun) yaml2go(wd, yamlpath, gopath string, assetfs *bool) error {
	//fmt.Println(yamlpath, "→", gopath)

	yamlbuf, err := ioutil.ReadFile(yamlpath)
	if err != nil {
		return err
	}

	var root yaml.Node

	if err := yaml.Unmarshal(yamlbuf, &root); err != nil {
		return err
	}

	gobuf := "func init() {\n"

	walker := &yamlwalker{
		br:       br,
		gobuf:    &gobuf,
		assetfs:  assetfs,
		imports:  map[string]string{},
		yamlpath: yamlpath,
		wd:       wd,
	}

	if err := walker.yamlwalk(&root); err != nil {
		return err
	}

	gobuf += "}\n"

	gobufhead := "package main\n\nimport (\n"
	for pkg, alias := range walker.imports {
		if pkg == alias || strings.HasSuffix(pkg, "/"+alias) {
			gobufhead += fmt.Sprintf("\t%#v\n", pkg)
		} else {
			gobufhead += fmt.Sprintf("\t%s %#v\n", alias, pkg)
		}
	}
	gobufhead += ")\n\n" + gobuf

	if err := ioutil.WriteFile(gopath, []byte(gobufhead), 0644); err != nil {
		return err
	}

	return nil
}

func (br *buildrun) yaml2struct(w *yamlwalker, v *yaml.Node, si interface{}) error {
	val := reflect.ValueOf(si)
	typ := val.Type()

	for typ.Kind() == reflect.Ptr {
		// zero out what it points to so we don't have danglings from the last yaml2struct call
		z := reflect.Zero(typ.Elem())
		val.Elem().Set(z)

		// dereference
		val = val.Elem()
		typ = val.Type()
	}

	Title := typ.Name()
	title := strings.ToLower(Title)

	fields := map[string]reflect.Value{}
	fieldtypes := map[string]reflect.StructField{}

	var (
		shortkeyk   string
		shortkeyf   reflect.Value
		shortkeyt   reflect.StructField
		shortvaluek string
		shortvaluef reflect.Value
		shortvaluet reflect.StructField
	)

	for i := 0; i < typ.NumField(); i++ {
		field := val.Field(i)
		ft := typ.Field(i)

		key := strings.ToLower(ft.Name)

		if tv, ok := ft.Tag.Lookup("khan"); ok {
			t, to := parseTag(tv)

			if t != "" {
				key = t
			}

			if to.Contains("shortkey") {
				shortkeyk = key
				shortkeyf = field
				shortkeyt = ft
			}

			if to.Contains("shortvalue") {
				shortvaluek = key
				shortvaluef = field
				shortvaluet = ft
			}

			if t == "-" {
				// Don't parse this struct field from the yaml
				continue
			}
		}

		fields[key] = field
		fieldtypes[key] = ft
	}

	source := fmt.Sprintf("%s:%d", w.yamlpath, v.Line)

	khanalias := w.addimport(khanpkgname, khanpkgalias)
	*w.gobuf += fmt.Sprintf("\t%s.AddFromSource(%#v, &%s.%s{", khanalias, source, khanalias, typ.Name())
	any := false
	alreadyset := map[string]bool{}

	if shortkeyk != "" && w.shortkey != "" {
		if alreadyset[shortkeyk] {
			return w.nodeErrorf(v, "%s %s set multiple times", Title, shortkeyk)
		}
		alreadyset[shortkeyk] = true

		f := shortkeyf
		ft := shortkeyt

		if !any {
			*w.gobuf += "\n"
			any = true
		}

		if !f.CanSet() {
			return w.nodeErrorf(v, "%s %s cannot be set", Title, shortkeyk)
		}

		if err := yaml2value(w, v, yaml.ScalarNode, w.shortkey, f); err != nil {
			return err
		}

		*w.gobuf += fmt.Sprintf("\t\t%s: %#v,\n", ft.Name, f.Interface())
	}

	if v.Kind == yaml.ScalarNode && shortvaluek != "" {
		if alreadyset[shortvaluek] {
			return w.nodeErrorf(v, "%s %s set multiple times", Title, shortvaluek)
		}
		alreadyset[shortvaluek] = true

		f := shortvaluef
		ft := shortvaluet

		if !any {
			*w.gobuf += "\n"
			any = true
		}

		if !f.CanSet() {
			return w.nodeErrorf(v, "%s %s cannot be set", Title, shortvaluek)
		}

		if err := yaml2value(w, v, yaml.ScalarNode, v.Value, f); err != nil {
			return err
		}

		*w.gobuf += fmt.Sprintf("\t\t%s: %#v,\n", ft.Name, f.Interface())
	} else if v.Kind == yaml.MappingNode {

		if len(v.Content)%2 != 0 {
			return w.nodeErrorf(v, "Odd sized YAML map")
		}

		for i := 0; i < len(v.Content); i += 2 {
			k := v.Content[i]
			v := v.Content[i+1]

			if k.Kind != yaml.ScalarNode {
				return w.nodeErrorf(k, "%s expected scalar map key: Got %s", Title, yamlkind(k.Kind))
			}

			param := k.Value

			f, ok := fields[k.Value]
			if !ok {
				return w.nodeErrorf(k, "Unknown %s parameter %#v", title, param)
			}
			ft := fieldtypes[k.Value]

			if alreadyset[k.Value] {
				return w.nodeErrorf(k, "%s %s set multiple times", Title, param)
			}
			alreadyset[k.Value] = true

			// TODO support arrays and structs

			if !any {
				*w.gobuf += "\n"
				any = true
			}

			if !f.CanSet() {
				return w.nodeErrorf(v, "%s %s cannot be set", Title, k.Value)
			}

			if err := yaml2value(w, v, v.Kind, v.Value, f); err != nil {
				return err
			}

			*w.gobuf += fmt.Sprintf("\t\t%s: %#v,\n", ft.Name, f.Interface())
		}

	} else {
		if shortvaluek == "" {
			return w.nodeErrorf(v, "Expected map: Got %s", yamlkind(v.Kind))
		} else {
			return w.nodeErrorf(v, "Expected map or a scalar %s: Got %s", yamlkind(v.Kind), shortvaluek)
		}
	}

	if any {
		*w.gobuf += "\t"
	}
	*w.gobuf += "})\n"

	// Validate struct
	siv, ok := si.(khan.Validator)
	if ok {
		if err := siv.Validate(); err != nil {
			return w.nodeErrorf(v, "%w", err)
		}
	}

	// Include static files into the go binary
	sif, ok := si.(khan.StaticFiler)
	if ok {
		files := sif.StaticFiles()
		for _, file := range files {
			br.staticfiles = append(br.staticfiles, file)
		}
	}

	return nil
}

func yaml2value(w *yamlwalker, v *yaml.Node, kind yaml.Kind, value string, dest reflect.Value) error {
	typ := dest.Type()

	// Special handling for this type: Parse as octal
	if typ == reflect.TypeOf(os.FileMode(0)) {
		if kind != yaml.ScalarNode {
			return w.nodeErrorf(v, "Expected scaler convertable to %s: Got %s", typ.Kind(), yamlkind(kind))
		}
		vi, err := strconv.ParseUint(value, 8, 32)
		if err != nil {
			return w.nodeErrorf(v, "Conversion from octal to uint32 failed: %w", err)
		}
		dest.SetUint(vi)
		return nil
	}

	// General type handling
	switch typ.Kind() {
	case reflect.String:
		if kind != yaml.ScalarNode {
			return w.nodeErrorf(v, "Expected scaler convertable to %s: Got %s", typ.Kind(), yamlkind(kind))
		}
		dest.SetString(value)
	case reflect.Int:
		if kind != yaml.ScalarNode {
			return w.nodeErrorf(v, "Expected scaler convertable to %s: Got %s", typ.Kind(), yamlkind(kind))
		}
		vi, err := strconv.Atoi(value)
		if err != nil {
			return w.nodeErrorf(v, "Conversion to int failed: %w", err)
		}
		dest.SetInt(int64(vi))
	case reflect.Uint32:
		if kind != yaml.ScalarNode {
			return w.nodeErrorf(v, "Expected scaler convertable to %s: Got %s", typ.Kind(), yamlkind(kind))
		}
		vi, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return w.nodeErrorf(v, "Conversion to uint32 failed: %w", err)
		}
		dest.SetUint(vi)
	case reflect.Bool:
		if kind != yaml.ScalarNode {
			return w.nodeErrorf(v, "Expected scaler convertable to %s: Got %s", typ.Kind(), yamlkind(kind))
		}
		vb, err := strconv.ParseBool(value)
		if err != nil {
			return w.nodeErrorf(v, "Conversion to boolean failed: %w", err)
		}
		dest.SetBool(vb)
	case reflect.Slice:
		sv := reflect.MakeSlice(typ, len(v.Content), len(v.Content))

		if kind != yaml.SequenceNode {
			// Special case: Empty scalar is allowed as empty list.
			if kind == yaml.ScalarNode && value == "" {
				dest.Set(sv)
				return nil
			}

			return w.nodeErrorf(v, "Expected array: Got %s", yamlkind(kind))
		}

		for i := 0; i < len(v.Content); i++ {
			vv := v.Content[i]
			rv := reflect.New(typ.Elem())
			if err := yaml2value(w, vv, vv.Kind, vv.Value, rv.Elem()); err != nil {
				return err
			}
			sv.Index(i).Set(rv.Elem())
		}

		dest.Set(sv)

	default:
		return w.nodeErrorf(v, "Unhandled type %s", typ.Kind())
	}
	return nil
}

func yamlsimplehandler(vv khan.Item) yamlhandler {
	return func(w *yamlwalker, v *yaml.Node) error {
		if err := w.br.yaml2struct(w, v, vv); err != nil {
			return err
		}
		return nil
	}
}
