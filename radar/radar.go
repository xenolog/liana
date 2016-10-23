package radar

import (
	"fmt"
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
	RR := fmt.Sprintf("Radar(%s):", if_name)
	if r.nic, err = net.InterfaceByName(if_name); err != nil {
		r.log.Warn("%s network interface not found.", RR)
		return
	}
	if addrs, err = r.nic.Addrs(); err != nil || 1 > len(addrs) {
		r.log.Warn("%s No ip addresses given.", RR)
		return
	}
	r.log.Debug("%s interface has %d addresses", RR, len(addrs))
	for i, a := range addrs {
		addr := a.String()
		r.log.Debug("%s %02d: interface has address '%s'", RR, i, addr)
		// if r.ipv4only && strings.Contains(addr, ":") { // is it a prohibited ipv6 address
		if strings.Contains(addr, ":") {
			continue
		} else {
			r.ipv4 = []byte(strings.Split(addr, "/")[0])
			break
		}
	}
	if len(r.ipv4) == 0 {
		r.log.Warn("%s No IPv4 addresses given.", RR)
		return
	} else {
		r.log.Info("%s address '%s' will be used while discovering", RR, string(r.ipv4))
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
