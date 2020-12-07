package main

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func yamlhandlerfile(w *yamlwalker, v *yaml.Node) error {
	if v.Kind != yaml.MappingNode {
		return w.nodeErrorf(v, "Expected map: Got %s", yamlkind(v.Kind))
	}

	if len(v.Content)%2 != 0 {
		return w.nodeErrorf(v, "Odd sized YAML map")
	}

	valid := map[string]bool{
		"path":    true,
		"content": true,
	}
	kv := map[string]string{}

	for i := 0; i < len(v.Content); i += 2 {
		k := v.Content[i]
		v := v.Content[i+1]
		if k.Kind != yaml.ScalarNode {
			return w.nodeErrorf(k, "Expected scalar map key: Got %s", yamlkind(k.Kind))
		}
		if v.Kind != yaml.ScalarNode {
			return w.nodeErrorf(v, "Expected scaler: Got %s", yamlkind(v.Kind))
		}
		if !valid[k.Value] {
			return w.nodeErrorf(k, "Invalid duck-yaml file parameter: %#v", k.Value)
		}
		kv[k.Value] = v.Value
	}

	alias := w.addimport(duckpkgname, duckpkgalias)
	*w.gobuf += fmt.Sprintf(`	%s.Add(&%s.File{
		Path: %#v,
		Content: %#v,
	})
`, alias, alias, kv["path"], kv["content"])

	return nil
}
