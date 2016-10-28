package registry

import (
	"github.com/xenolog/liana/config"
	//"github.com/xenolog/liana/ct"
	"errors"
	"fmt"
	"sync"
	// "time"
)

type Endpoint string

type Node struct {
	sync.Mutex
	Name      string
	endpoints []Endpoint
}
type NodesList map[string]Node

type Registry struct {
	sync.Mutex
	cfg   *config.Config
	Nodes NodesList
}

var local_registry *Registry

func (r *Registry) nodeExists(node_name string, unlock bool) bool {
	// non-public method, which allows does not remove mutex unlock
	// for operate with registry and unlock mutex manually later
	rv := true
	r.Lock()
	if _, ok := r.Nodes[node_name]; !ok {
		rv = false
	}
	if unlock {
		r.Unlock()
	}
	return rv
}

func (r *Registry) NodeExists(node_name string) bool {
	return r.nodeExists(node_name, true)
}

func (r *Registry) AddNode(node_name string) error {
	var err error
	if !r.nodeExists(node_name, false) {
		r.Nodes[node_name] = Node{Name: node_name}
		r.Unlock()
		err = nil
	} else {
		r.Unlock()
		err = errors.New(fmt.Sprintf("Node '%s' alredy exists", node_name))
	}
	return err
}

func (r *Registry) RemoveNode(node_name string) error {
	var err error
	if r.nodeExists(node_name, false) {
		delete(r.Nodes, node_name)
		r.Unlock()
		err = nil
	} else {
		r.Unlock()
		err = errors.New(fmt.Sprintf("Node '%s' does not exists", node_name))
	}
	return err
}

func NewRegistry(cfg *config.Config) *Registry {
	if cfg != nil {
		local_registry.cfg = cfg
	}
	return local_registry
}

func init() {
	local_registry = new(Registry)
	local_registry.Nodes = make(NodesList)
}
