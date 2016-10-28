package registry_test

import (
	"github.com/xenolog/liana/registry"
	"testing"
)

var local_registry *registry.Registry

func init() {
	local_registry = registry.NewRegistry(nil)
	local_registry.Nodes["xxx1"] = registry.Node{
		Name: "xxx1",
	}
	local_registry.Nodes["xxx2"] = registry.Node{
		Name: "xxx2",
	}
}

func Test_NodeExists(t *testing.T) {
	// var nodes ct.NodesList
	node_name := "xxx3"
	if local_registry.NodeExists(node_name) {
		t.Errorf("Node '%s' found, but not exists", node_name)
	}
	node_name = "xxx2"
	if !local_registry.NodeExists(node_name) {
		t.Errorf("Node '%s' not found, but exists", node_name)
	}
}

func Test_AddNode(t *testing.T) {
	var (
		err error
	)
	node_name := "xxx3"
	if err = local_registry.AddNode(node_name); err != nil {
		t.Errorf("Node '%s' added, non-nil error returned: %s", node_name, err)
	}
	if !local_registry.NodeExists(node_name) {
		t.Errorf("Node '%s' added, but not found", node_name)
	}
	// add diplicated node
	if err = local_registry.AddNode(node_name); err == nil {
		t.Errorf("Duplicate node '%s' added, error not returned", node_name)
	}
}

func Test_RemoveNode(t *testing.T) {
	var (
		err error
	)
	node_name := "xxx3"
	if err = local_registry.RemoveNode(node_name); err != nil {
		t.Errorf("Node '%s' removed, non-nil error returned: %s", node_name, err)
	}
	if local_registry.NodeExists(node_name) {
		t.Errorf("Node '%s' removed, but found in the nodes list", node_name)
	}
}
