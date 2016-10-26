package discovery

import (
	"github.com/xenolog/liana/config"
	"sync"
	"time"
)

const RADARsFANOUT_BUFFER = 64

type Discovery struct {
	sync.Mutex
	cfg            *config.Config
	interfaces     []string
	passwd         string
	radars         map[string]*Radar
	radarsFanout   RadarFanout   // nil if no
	stopRadars     chan struct{} // Every Radar and Responder
	stopResponders chan struct{} // gorutine listen this chan for stop itself

}

var host_discovery *Discovery

func (d *Discovery) RemoveRadar(if_name string) {
	if _, ok := d.radars[if_name]; !ok {
		return
	}
	d.cfg.Log.Debug("Destroy Radar for '%s'", if_name)
	delete(d.radars, if_name)
}

func (d *Discovery) AddRadar(if_name string) {
	if _, ok := d.radars[if_name]; ok {
		return
	}
	d.cfg.Log.Debug("Try to create Radar for '%s'", if_name)
	if new_radar := NewRadar(d.cfg, if_name, d.radarsFanout); new_radar != nil {
		d.radars[if_name] = new_radar
		go d.radars[if_name].Run()
	}
}

func (d *Discovery) radarRunner() {
	if d.radarsFanout != nil {
		return
	}
	d.radarsFanout = make(chan string, RADARsFANOUT_BUFFER)
	for {
		for _, if_name := range d.interfaces {
			if if_name[len(if_name)-1] == '*' {
				//todo: handle an interface wildcard
			} else {
				if _, ok := d.radars[if_name]; !ok {
					d.AddRadar(if_name)
				}
			}
		}
		// handle signals from Liana core
		select {
		case <-time.After(15 * time.Second):
			d.cfg.Log.Debug("Radar RUNNER timeout")
		case <-d.stopRadars: // stop all radars
			for _, r := range d.radars {
				r.stopRadar <- struct{}{}
				//r.stopResponder <- struct{}{}
			}
			close(d.radarsFanout)
			break
		}
	}
}

func (d *Discovery) responderRunner() {
	// 	for {
	// 		select {
	// 		case msg := <-radarsFanout:
	// 			// process radars, start responders
	// 			switch {
	// 			case msg[0] == '+':
	// 				d.cfg.Log.Debug("Message from Radar: '%s'", msg)
	// 				// Responder for Radar should be run
	// 			case msg[0] == '+':
	// 				d.cfg.Log.Debug("Message from Radar: '%s'", msg)
	// 				// Responder for Radar should be died
	// 			case msg[0] == '!':
	// 				d.cfg.Log.Debug("Radar was died: '%s'", msg)
	// 				d.RemoveRadar(msg[1:])
	// 			}
	// 		case <-time.After(5 * time.Second):
	// 			// Do nothing, just timeout happens (for the future usage)
	// 			d.cfg.Log.Debug("Discoverer RUN timeout")
	// 		}
	// 	}
}

func (d *Discovery) Run() {
	go d.radarRunner()
	go d.responderRunner()
}

func NewDiscovery(cfg *config.Config, ifaces *[]string) *Discovery {
	if cfg != nil {
		host_discovery.cfg = cfg
	}
	if *ifaces != nil {
		host_discovery.interfaces = *ifaces
	}
	host_discovery.radars = make(map[string]*Radar)
	return host_discovery
}

func init() {
	host_discovery = new(Discovery)
	host_discovery.radars = make(map[string]*Radar)
}
