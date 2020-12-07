package main

import (
	"github.com/yobert/duck"
	"gopkg.in/yaml.v3"
)

func yamlhandlergroup(w *yamlwalker, v *yaml.Node) error {
	var g duck.Group
	if err := yaml2struct(w, v, &g); err != nil {
		return err
	}
	return nil
}
