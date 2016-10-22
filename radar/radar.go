package radar

import (
	"gopkg.in/xenolog/go-tiny-logger.v1"
	"net"
	"strings"
)

type Radar struct {
	ipv4only bool
	log      *logger.Logger
	nic      *net.Interface
	ipv4     net.IP
	udpport  int
}

func (r *Radar) Run(if_name string, passwd string) {
	var (
		err error
		// addr  net.Addr
		addrs []net.Addr
	)
	if r.nic, err = net.InterfaceByName(if_name); err != nil {
		r.log.Warn("Radar: network interface '%s' not found.", if_name)
		return
	}
	if addrs, err = r.nic.Addrs(); err != nil || 1 > len(addrs) {
		r.log.Warn("Radar: No ip addresses for interface '%s' given.", if_name)
		return
	}
	r.log.Debug("Radar: interface '%s' has %d addresses", if_name, len(addrs))
	for i, a := range addrs {
		addr := a.String()
		r.log.Debug("Radar: %02d: interface '%s' has address '%s'", i, if_name, addr)
		// if r.ipv4only && strings.Contains(addr, ":") { // is it a prohibited ipv6 address
		if strings.Contains(addr, ":") {
			continue
		} else {
			r.ipv4 = []byte(strings.Split(addr, "/")[0])
			break
		}
	}
	if len(r.ipv4) == 0 {
		r.log.Warn("Radar: No IPv4 addresses for interface '%s' given.", if_name)
		return
	} else {
		r.log.Debug("Radar: interface '%s' will use address '%s' while discovering", if_name, string(r.ipv4))
	}
}

///
func NewRadar(l *logger.Logger, flags []string) *Radar {
	r := new(Radar)
	r.log = l
	for _, v := range flags {
		switch v {
		case "ipv4only":
			r.ipv4only = true
		}
	}

	return r
}
