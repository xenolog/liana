package discovery

import (
	"encoding/json"
	"fmt"
	"github.com/xenolog/liana/config"
	"net"
	"sync"
)

const READ_CHAN_BUFF = 32

type ResponderFanout chan string

type Responder struct {
	sync.Mutex
	cfg           *config.Config
	nic_name      string
	self          *net.UDPAddr
	stopResponder chan struct{} // any message from this chan is a command to destroy an Radar
	fanout        ResponderFanout
}
type ResponderMap map[string]*Responder

func (r *Responder) DeCryptBeacon(txt []byte) ([]byte, error) {
	// beacon should be crypted in the future
	return txt, nil
}

// first byte is a action:
// '+'     -- anounce received
// '!'     -- destroy Responder's data structures
func (r *Responder) fanoutMsg(msg string) {
	op := msg[0]
	switch op {
	case '+':
		r.fanout <- fmt.Sprintf("%s", msg)
	case '!':
		r.fanout <- fmt.Sprintf("%s%s", op, r.nic_name)
	}
}

func (r *Responder) Run() {
	var (
		err            error
		nic            *net.Interface
		listen_ip_port *net.UDPAddr
		in_mcast_conn  *net.UDPConn
		beacon, buff   []byte
		read_buffer    []byte
		ConnInfo       Beacon
		from_network   chan []byte
	)
	RR := fmt.Sprintf("Responder(%s):", r.nic_name)
	r.cfg.Log.Debug("%s starting...", RR)
	if nic, err = net.InterfaceByName(r.nic_name); err != nil {
		r.cfg.Log.Error("%s Interface search by name throw error: %s", RR, err)
		return
	}
	listen_ip_port, _ = net.ResolveUDPAddr("udp", r.cfg.McastDestination)
	if in_mcast_conn, err = net.ListenMulticastUDP("udp", nic, listen_ip_port); err != nil {
		r.cfg.Log.Error("%s Can't start mcast listening on interface '%s', for '%s': %s", RR, r.nic_name, listen_ip_port, err)
		return
	}
	in_mcast_conn.SetReadBuffer(1048576)
	from_network = make(chan []byte, READ_CHAN_BUFF)
	read_buffer = make([]byte, 1500) // about MTU of interface
	for {
		// detach read from network
		go func(c *net.UDPConn, data []byte, self *net.UDPAddr, ch chan []byte) {
			var (
				income *net.UDPAddr
				err    error
			)
			for {
				//TODO: Handle timeout for ReadFromUDP
				// ReadFromUDP can be made to time out and return
				// an error with Timeout() == true after a fixed time limit;
				// see SetDeadline and SetReadDeadline.
				if _, income, err = c.ReadFromUDP(data); err != nil {
					r.cfg.Log.Error("%s error while mcast packer read: %s", RR, err)
					data = nil
				}
				if fmt.Sprintf("%s", self.IP) == fmt.Sprintf("%s", income.IP) {
					// filter self-made packages
					continue
				} else {
					// r.cfg.Log.Debug("%s XX self='%s' income='%s'", RR, fmt.Sprintf("%s", self.IP), fmt.Sprintf("%s", income.IP))
					// r.cfg.Log.Debug("%s XX read from network: '%s'", RR, data)
					ch <- data
					break
				}
			}
		}(in_mcast_conn, read_buffer, r.self, from_network)

		select {
		case <-r.stopResponder:
			// signal from control plane is time to die
			close(r.stopResponder)
			r.fanoutMsg("!")
			break
		case buff = <-from_network:
			// do nothing here, data from network will be processed below
		}
		// if buff != []byte{} {
		if buff != nil {
			if beacon, err = r.DeCryptBeacon(buff); err != nil {
				r.cfg.Log.Error("%s error while beacon decription: %s", RR, err)
				continue
			} else {
				r.cfg.Log.Debug("%s received JSON: %s", RR, beacon)
			}

			if err := json.Unmarshal(beacon, &ConnInfo); err != nil {
				r.cfg.Log.Error("%s error while beacon parse: %s", RR, err)
				continue
			}

			r.fanoutMsg(fmt.Sprintf("+%s:%s", ConnInfo.Name, ConnInfo.Endpoints))
		}
	}
}

///
func NewResponder(cfg *config.Config, if_name string, self_source string, ch ResponderFanout) *Responder {
	new_responder := new(Responder)
	new_responder.cfg = cfg
	new_responder.fanout = ch
	new_responder.stopResponder = make(chan struct{}, 1)
	new_responder.nic_name = if_name
	new_responder.self, _ = net.ResolveUDPAddr("udp", self_source)
	cfg.Log.Debug("Responder for '%s' created", if_name)
	return new_responder
}
