package keval

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
	//"github.com/davecgh/go-spew/spew"
)

func builtin_from_yaml(fpath string) interface{} {
	fh, err := os.Open(fpath)
	if err != nil {
		panic(err)
	}
	defer fh.Close()

	var r yaml.Node

	dec := yaml.NewDecoder(fh)

	if err := dec.Decode(&r); err != nil {
		panic(err)
	}

	return builtin_from_yaml_node(&r)
}

// Convert yaml types into simple go types.
func builtin_from_yaml_node(node *yaml.Node) interface{} {
	if node == nil {
		return nil
	}

	if node.Kind == yaml.SequenceNode {
		r := make([]interface{}, len(node.Content))
		for i, v := range node.Content {
			r[i] = builtin_from_yaml_node(v)
		}
		return r
	}

	if node.Kind == yaml.MappingNode {
		r := make(map[string]interface{})
		for i := 0; i < len(node.Content); i += 2 {
			k := node.Content[i]
			v := node.Content[i+1]
			if k != nil {
				r[k.Value] = builtin_from_yaml_node(v)
			}
		}
		return r
	}

	if node.Kind == yaml.DocumentNode {
		return builtin_from_yaml_node(node.Content[0])
	}

	if node.Kind == yaml.ScalarNode {
		return node.Value
	}

	// This might endless loop if the alias graph has cycles.... TODO
	if node.Kind == yaml.AliasNode {
		return builtin_from_yaml_node(node.Alias)
	}

	panic(fmt.Sprintf("Unhandled yaml node kind %d", node.Kind))
}

func (m *Machine) EvalYaml(fpath string) error {

}
