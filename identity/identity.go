package identity

import (
	"gopkg.in/xenolog/go-tiny-logger.v1"
	"os"
	"sync"
)

type HostIdentity struct {
	sync.RWMutex
	log        *logger.Logger
	configured bool
	hostname   string
}

var main_identity *HostIdentity

func (hi *HostIdentity) GetHostname() string {
	return hi.hostname
}

func (hi *HostIdentity) FetchSystemHostname() {
	var err error
	hi.Lock()
	defer hi.Unlock()
	hi.hostname, err = os.Hostname()
	if err != nil {
		hi.log.Fail("!!! Can't understand hostname: %s", err)
		os.Exit(1)
	}
}

func (hi *HostIdentity) Run() {
	// do nothing now.
	// todo: hostname and keys change monitoring should be running here
	return
}

///
func New(l *logger.Logger) *HostIdentity {
	if !main_identity.configured {
		// initial configuration
		main_identity.FetchSystemHostname()
		main_identity.configured = true
	}
	return main_identity
}

func init() {
	main_identity = new(HostIdentity)
}
