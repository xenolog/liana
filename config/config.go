package config

import (
	"github.com/xenolog/liana/identity"
	"gopkg.in/xenolog/go-tiny-logger.v1"
	"sync"
	"time"
)

type Config struct {
	sync.Mutex
	Log              *logger.Logger
	McastDestination string
	McastInterval    time.Duration
	ListenPort       int
	Identity         *identity.HostIdentity
}

var host_config *Config

// func (hi *Config) GetHostname() string {
// 	return hi.hostname
// }

///
func New() *Config {
	return host_config
}

func init() {
	host_config = new(Config)
}
