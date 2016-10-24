package discovery

import (
	"github.com/xenolog/liana/config"
	"sync"
)

type Discovery struct {
	sync.RWMutex
	cfg *config.Config
}

var host_discovery *Discovery

func (d *Discovery) Run() {
}

func New(cfg *config.Config) *Discovery {
	host_discovery.cfg = cfg
	return host_discovery
}

func init() {
	host_discovery = new(Discovery)
}
