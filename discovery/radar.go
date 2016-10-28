package discovery

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

type RadarFanout chan string

type Radar struct {
	sync.Mutex
	ipv4only  bool
	cfg       *config.Config
	nic       *net.Interface
	ipv4      string
	udpport   int
	stopRadar chan struct{} // any message from this chan is a command to destroy an Radar
	fanout    RadarFanout
}
type RadarMap map[string]*Radar

type Beacon struct {
	Version   string `json:"version"`
	Name      string `json:"name"`
	Endpoints string `json:"endpoint"`
}

func (r *Radar) Sleep() {
	time.Sleep(r.cfg.McastInterval * time.Second)
}

func (r *Radar) CryptBeacon(txt []byte) []byte {
	// beacon should be crypted in the future
	return txt
}

// first byte is a action:
// '+','-' -- add/remove Responder for the interface
// '!'     -- destroy Radar and Responder
func (r *Radar) fanoutMsg(op string) {
	var msg string
	switch op {
	case "+", "-":
		msg = fmt.Sprintf("%s%s:%s:%d", op, r.nic.Name, r.ipv4, r.cfg.ListenPort)
	case "!":
		msg = fmt.Sprintf("%s%s", op, r.nic.Name)
	}
	r.cfg.Log.Debug("Radar(%s): fanout message: '%s'", r.nic.Name, msg)
	r.fanout <- msg
}

func (r *Radar) Run() {
	var (
		err                      error
		addrs                    []net.Addr
		src_ip_port, dst_ip_port *net.UDPAddr
		out_mcast_conn           *net.UDPConn
	)
	RR := fmt.Sprintf("Radar(%s):", r.nic.Name)
	for {
		// This is a monitoring for NIC alive or dead
		if _, err = net.InterfaceByName(r.nic.Name); err != nil {
			r.cfg.Log.Debug("%s NIC was destroyed, Radar will be distroyed", RR)
			r.fanoutMsg("-")
			r.fanoutMsg("!")
			break
		}
		// Calculate source address for beacon
		if addrs, err = r.nic.Addrs(); err != nil || 1 > len(addrs) {
			r.cfg.Log.Warn("%s No ip addresses given.", RR)
			r.fanoutMsg("-")
			r.Sleep()
			continue
		}
		r.cfg.Log.Debug("%s interface has %d addresses", RR, len(addrs))
		for i, a := range addrs {
			addr := a.String()
			r.cfg.Log.Debug("%s %02d: interface has address '%s'", RR, i, addr)
			// if r.ipv4only && strings.Contains(addr, ":") { // is it a prohibited ipv6 address
			if strings.Contains(addr, ":") {
				continue // it's a IPv6 address, see next address
			} else {
				r.ipv4 = strings.Split(addr, "/")[0]
				break
			}
		}
		if r.ipv4 == "" {
			r.cfg.Log.Warn("%s No IPv4 addresses given.", RR)
			r.fanoutMsg("-")
			r.Sleep()
			continue
		} else {
			r.cfg.Log.Debug("%s address '%s' will be used into discovery beacon", RR, r.ipv4)
		}
		r.cfg.Log.Debug("%s destination is '%s'", RR, r.cfg.McastDestination)
		src_ip_port, _ = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%s", r.ipv4, r.cfg.ListenPort))
		dst_ip_port, _ = net.ResolveUDPAddr("udp", r.cfg.McastDestination)
		if out_mcast_conn, err = net.DialUDP("udp", src_ip_port, dst_ip_port); err != nil {
			r.cfg.Log.Error("%s Can't create outbound socket for beacon: %s", RR, err)
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
			r.cfg.Log.Error("%s Can't send beacon: %s", RR, err)
			r.Sleep()
			continue
		}
		out_mcast_conn.Close()
		// sleep before next beacon
		select {
		case <-time.After(r.cfg.McastInterval * time.Second):
			// just timeout
		case <-r.stopRadar:
			close(r.stopRadar)
			r.fanoutMsg("!")
			break
		}
	}
}

///
func NewRadar(cfg *config.Config, if_name string, ch RadarFanout) *Radar {
	var (
		err error
		nic *net.Interface
	)
	if nic, err = net.InterfaceByName(if_name); err != nil {
		//cfg.Log.Debug("Interface search by name throw error: %s", err)
		return nil
	}
	new_radar := new(Radar)
	new_radar.cfg = cfg
	new_radar.fanout = ch
	new_radar.stopRadar = make(chan struct{}, 1)
	new_radar.nic = nic
	cfg.Log.Debug("Radar for '%s' created", if_name)
	return new_radar
}
