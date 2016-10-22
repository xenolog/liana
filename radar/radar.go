package radar

import (
	"gopkg.in/xenolog/go-tiny-logger.v1"
	// "net"
	// "net"
	// "strings"
)

type Radar struct {
	Log *logger.Logger
}

func (r *Radar) Run(iface string, passwd string) {
	return
}

///
func NewRadar(l *logger.Logger) *Radar {
	r := new(Radar)
	r.Log = l
	return r
}
