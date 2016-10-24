package radar

import (
	"encoding/json"
	"fmt"
	"github.com/xenolog/liana/config"
	"net"
	"strings"
	"sync"
	"time"
)

const version = "0.0.1"

type Radar struct {
	sync.RWMutex
	ipv4only bool
	flags    []string
	cfg      *config.Config
	nic      *net.Interface
	ipv4     string
	udpport  int
}

type Beacon struct {
	Version   string `json:"version"`
	Name      string `json:"name"`
	Endpoints string `json:"endpoint"`
}

func (r *Radar) AddFlag(flag string) {
	r.Lock()
	defer r.Unlock()
	r.flags = append(r.flags, flag)
	// process some important flags
	for _, v := range r.flags {
		switch v {
		case "ipv4only":
			r.ipv4only = true
		}
	}
}

func (r *Radar) Sleep() {
	time.Sleep(r.cfg.McastInterval * time.Second)
}

func (r *Radar) CryptBeacon(txt []byte) []byte {
	// beacon should be crypted in the future
	return txt
}

func (r *Radar) Run(if_name string, passwd string) {
	var (
		err                      error
		addrs                    []net.Addr
		src_ip_port, dst_ip_port *net.UDPAddr
		out_mcast_conn           *net.UDPConn
	)
	RR := fmt.Sprintf("Radar(%s):", if_name)
	if r.nic, err = net.InterfaceByName(if_name); err != nil {
		r.cfg.Log.Warn("%s network interface not found.", RR)
		return
	}
	for {
		// Calculate source address for beacon
		if addrs, err = r.nic.Addrs(); err != nil || 1 > len(addrs) {
			r.cfg.Log.Warn("%s No ip addresses given.", RR)
			r.Sleep()
			continue
		}
		r.cfg.Log.Debug("%s interface has %d addresses", RR, len(addrs))
		for i, a := range addrs {
			addr := a.String()
			r.cfg.Log.Debug("%s %02d: interface has address '%s'", RR, i, addr)
			// if r.ipv4only && strings.Contains(addr, ":") { // is it a prohibited ipv6 address
			if strings.Contains(addr, ":") {
				continue
			} else {
				r.ipv4 = strings.Split(addr, "/")[0]
				break
			}
		}
		if r.ipv4 == "" {
			r.cfg.Log.Warn("%s No IPv4 addresses given.", RR)
			r.Sleep()
			continue
		} else {
			r.cfg.Log.Debug("%s address '%s' will be used into discovery beacon", RR, r.ipv4)
		}
		r.cfg.Log.Debug("%s destination is '%s'", RR, r.cfg.McastDestination)
		src_ip_port, _ = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%s", r.ipv4, r.cfg.ListenPort))
		dst_ip_port, _ = net.ResolveUDPAddr("udp", r.cfg.McastDestination)
		if out_mcast_conn, err = net.DialUDP("udp", src_ip_port, dst_ip_port); err != nil {
			r.cfg.Log.Error("Can't create outbound socket for beacon: %s", err)
			r.Sleep()
			continue
		}
		beacon, _ := json.Marshal(Beacon{
			Version:   version,
			Name:      r.cfg.Identity.GetHostname(),
			Endpoints: fmt.Sprintf("%s:%d", r.ipv4, r.cfg.ListenPort),
		})
		r.cfg.Log.Debug("%s beacon '%s'", RR, beacon)
		if _, err = out_mcast_conn.Write(r.CryptBeacon(beacon)); err != nil {
			r.cfg.Log.Error("Can't send beacon: %s", err)
			r.Sleep()
			continue
		}
		out_mcast_conn.Close()
		// sleep before next beacon
		r.Sleep()
	}
}

///
func NewRadar(cfg *config.Config) *Radar {
	r := new(Radar)
	r.cfg = cfg
	return r
}
