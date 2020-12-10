package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-bindata/go-bindata/v3"
	"github.com/yobert/duck"
	"gopkg.in/yaml.v3"
)

const (
	duckpkgname  = "github.com/yobert/duck"
	duckpkgalias = "duck"
)

type yamlwalker struct {
	gobuf    *string
	imports  map[string]string
	yamlpath string
	wd       string
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
	"file":  yamlsimplehandler(&duck.File{}),
	"group": yamlsimplehandler(&duck.Group{}),
	"user":  yamlsimplehandler(&duck.User{}),
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

			h, ok := yamlhandlers[k.Value]
			if !ok {
				return w.nodeErrorf(k, "Invalid duck-yaml type %#v", k.Value)
			}

			if err := h(w, v); err != nil {
				return err
			}
		}
		return nil
	}
	return w.nodeErrorf(node, "Expected array or map: Got %s", yamlkind(node.Kind))
}

func yaml2go(wd, yamlpath, gopath string) error {
	fmt.Println(yamlpath, "→", gopath)

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
		gobuf:    &gobuf,
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

func yaml2struct(w *yamlwalker, v *yaml.Node, si interface{}) error {
	if v.Kind != yaml.MappingNode {
		return w.nodeErrorf(v, "Expected map: Got %s", yamlkind(v.Kind))
	}

	if len(v.Content)%2 != 0 {
		return w.nodeErrorf(v, "Odd sized YAML map")
	}

	val := reflect.ValueOf(si)
	typ := val.Type()
	if typ.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = val.Type()
	}

	Title := typ.Name()
	title := strings.ToLower(Title)

	fields := map[string]reflect.Value{}
	fieldtypes := map[string]reflect.StructField{}
	for i := 0; i < typ.NumField(); i++ {
		field := val.Field(i)
		ft := typ.Field(i)
		if alias, ok := ft.Tag.Lookup("duck"); ok {
			if alias == "" {
				// disabled
			} else {
				fields[alias] = field
				fieldtypes[alias] = ft
			}
		} else {
			fields[strings.ToLower(ft.Name)] = field
			fieldtypes[strings.ToLower(ft.Name)] = ft
		}
	}

	duckalias := w.addimport(duckpkgname, duckpkgalias)
	*w.gobuf += fmt.Sprintf("\t%s.Add(&%s.%s{", duckalias, duckalias, typ.Name())
	any := false
	alreadyset := map[string]bool{}

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

		if err := yaml2value(w, v, f); err != nil {
			return err
		}

		*w.gobuf += fmt.Sprintf("\t\t%s: %#v,\n", ft.Name, f.Interface())
	}

	if any {
		*w.gobuf += "\t"
	}
	*w.gobuf += "})\n"

	// Validate struct
	siv, ok := si.(duck.Validator)
	if ok {
		if err := siv.Validate(); err != nil {
			return w.nodeErrorf(v, "%w", err)
		}
	}

	// Include static files into the go binary
	sif, ok := si.(duck.StaticFiler)
	if ok {
		files := sif.StaticFiles()
		for _, file := range files {
			file = filepath.Clean(file)
			c := bindata.NewConfig()
			c.Input = append(c.Input, bindata.InputConfig{
				Path:      file,
				Recursive: false,
			})
			c.Package = "main"

			gfn := strings.Replace(file, "/", "_", -1)
			c.Output = w.wd + "/go_bindata_static_file_" + gfn + ".go"
			fmt.Println(file, "→", c.Output)
			if err := bindata.Translate(c); err != nil {
				return err
			}
		}
	}

	return nil
}

func yaml2value(w *yamlwalker, v *yaml.Node, dest reflect.Value) error {
	typ := dest.Type()
	switch typ.Kind() {
	case reflect.String:
		if v.Kind != yaml.ScalarNode {
			return w.nodeErrorf(v, "Expected scaler convertable to %s: Got %s", typ.Kind(), yamlkind(v.Kind))
		}
		dest.SetString(v.Value)
	case reflect.Int:
		if v.Kind != yaml.ScalarNode {
			return w.nodeErrorf(v, "Expected scaler convertable to %s: Got %s", typ.Kind(), yamlkind(v.Kind))
		}
		vi, err := strconv.Atoi(v.Value)
		if err != nil {
			return w.nodeErrorf(v, "Conversion to integer failed: %w", err)
		}
		dest.SetInt(int64(vi))
	case reflect.Bool:
		if v.Kind != yaml.ScalarNode {
			return w.nodeErrorf(v, "Expected scaler convertable to %s: Got %s", typ.Kind(), yamlkind(v.Kind))
		}
		vb, err := strconv.ParseBool(v.Value)
		if err != nil {
			return w.nodeErrorf(v, "Conversion to boolean failed: %w", err)
		}
		dest.SetBool(vb)
	case reflect.Slice:
		sv := reflect.MakeSlice(typ, len(v.Content), len(v.Content))

		if v.Kind != yaml.SequenceNode {
			// Special case: Empty scalar is allowed as empty list.
			if v.Kind == yaml.ScalarNode && v.Value == "" {
				dest.Set(sv)
				return nil
			}

			return w.nodeErrorf(v, "Expected array: Got %s", yamlkind(v.Kind))
		}

		for i := 0; i < len(v.Content); i++ {
			vv := v.Content[i]
			rv := reflect.New(typ.Elem())
			if err := yaml2value(w, vv, rv.Elem()); err != nil {
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

func yamlsimplehandler(vv interface{}) yamlhandler {
	return func(w *yamlwalker, v *yaml.Node) error {
		if err := yaml2struct(w, v, vv); err != nil {
			return err
		}
		return nil
	}
}
