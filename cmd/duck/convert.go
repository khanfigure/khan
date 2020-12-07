package main

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"

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
	"file":  yamlhandlerfile,
	"group": yamlhandlergroup,
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

func yaml2go(yamlpath, gopath string) error {
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
		gobuf:    &gobuf,
		imports:  map[string]string{},
		yamlpath: yamlpath,
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
		if v.Kind != yaml.ScalarNode {
			return w.nodeErrorf(v, "%s %s expected scaler: Got %s", Title, param, yamlkind(v.Kind))
		}

		if !any {
			*w.gobuf += "\n"
			any = true
		}

		if !f.CanSet() {
			return w.nodeErrorf(v, "%s %s cannot be set", Title, k.Value)
		}
		switch ft.Type.Kind() {
		case reflect.String:
			f.SetString(v.Value)
		case reflect.Int:
			vi, err := strconv.Atoi(v.Value)
			if err != nil {
				return w.nodeErrorf(v, "%s %s conversion to integer failed: %w", Title, param, err)
			}
			f.SetInt(int64(vi))
		default:
			return w.nodeErrorf(v, "%s %s has unhandled type %s", Title, param, ft.Type.Kind())
		}

		*w.gobuf += fmt.Sprintf("\t\t%s: %#v,\n", ft.Name, f.Interface())
	}

	if any {
		*w.gobuf += "\t"
	}
	*w.gobuf += "})\n"

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
