package discovery

import (
	"errors"
	"fmt"
	"github.com/xenolog/liana/config"
	"strings"
	"sync"
	"time"
)

const RADARS_FANOUT_BUFFER = 64

type Discovery struct {
	sync.Mutex
	cfg              *config.Config
	interfaces       []string
	passwd           string
	radars           RadarMap
	responders       ResponderMap
	radarsFanout     RadarFanout
	respondersFanout ResponderFanout
	stopRadars       chan struct{} // Every Radar and Responder
	stopResponders   chan struct{} // gorutine listen this chan for stop itself
	RadarsAlive      bool
	RespondersAlive  bool
}

var host_discovery *Discovery

// RESPONDER

func (d *Discovery) GetResponder(if_name string) (*Responder, error) {
	var err error
	d.Lock()
	defer d.Unlock()
	resp, ok := d.responders[if_name]
	if ok {
		err = nil
	} else {
		err = errors.New(fmt.Sprintf("Responder '%s' not found", if_name))
	}
	return resp, err
}

func (d *Discovery) responderExists(if_name string, unlock bool) bool {
	d.Lock()
	_, ok := d.responders[if_name]
	if unlock {
		d.Unlock()
	}
	return ok
}

func (d *Discovery) RespondersExists(if_name string) bool {
	return d.responderExists(if_name, true)
}

func (d *Discovery) RemoveResponder(if_name string) {
	if d.responderExists(if_name, false) {
		d.cfg.Log.Debug("Destroy responder for '%s'", if_name)
		delete(d.responders, if_name)
	}
	d.Unlock()
}

func (d *Discovery) AddResponder(if_name string) {
	if !d.responderExists(if_name, false) {
		d.cfg.Log.Debug("Try to create Responder for '%s'", if_name)
		if new_responder := NewResponder(d.cfg, if_name, d.respondersFanout); new_responder != nil {
			d.responders[if_name] = new_responder
			go d.responders[if_name].Run()
		}
		d.RespondersAlive = true
	}
	d.Unlock()
}

func (d *Discovery) responderRunner() {
	for {
		select {
		case msg := <-d.respondersFanout:
			switch {
			case msg[0] == '+':
				d.cfg.Log.Debug("XXX: '%s'", msg[1:])
			case msg[0] == '!':
				d.cfg.Log.Debug("Responder was died: '%s'", msg)
				if resp, err := d.GetResponder(msg[1:]); err == nil {
					close(resp.stopResponder)
					d.RemoveResponder(msg[1:])
				} else {
					d.cfg.Log.Error("%s", err)
				}
			}
		case msg := <-d.radarsFanout:
			// process radars, start responders
			switch {
			case msg[0] == '+':
				d.cfg.Log.Debug("Message from Radar: '%s'", msg)
				// Responder for Radar should be run
				resp_info := strings.SplitN(msg[1:], ":", 2)
				d.AddResponder(resp_info[0])
			case msg[0] == '-':
				d.cfg.Log.Debug("Message from Radar: '%s'", msg)
				// Responder for Radar should be died
				resp_info := strings.SplitN(msg[1:], ":", 2)
				if resp, err := d.GetResponder(resp_info[0]); err == nil {
					resp.stopResponder <- struct{}{}
				} else {
					d.cfg.Log.Error("%s", err)
				}
			case msg[0] == '!':
				d.cfg.Log.Debug("Radar was died: '%s'", msg)
				d.RemoveRadar(msg[1:])
			}
		//Do not remove, leave for debug purpose
		case <-time.After(5 * time.Second):
			// Do nothing, just timeout happens (for the future usage)
			d.cfg.Log.Debug("Discoverer RUN timeout")
		}
	}
}

// RADAR:

func (d *Discovery) radarExists(if_name string, unlock bool) bool {
	d.Lock()
	_, ok := d.radars[if_name]
	if unlock {
		d.Unlock()
	}
	return ok
}

func (d *Discovery) RadarExists(if_name string) bool {
	return d.radarExists(if_name, true)
}

func (d *Discovery) RemoveRadar(if_name string) {
	if d.radarExists(if_name, false) {
		d.cfg.Log.Debug("Destroy Radar for '%s'", if_name)
		delete(d.radars, if_name)
	}
	d.Unlock()
}

func (d *Discovery) AddRadar(if_name string) {
	if !d.radarExists(if_name, false) {
		d.cfg.Log.Debug("Try to create Radar for '%s'", if_name)
		if new_radar := NewRadar(d.cfg, if_name, d.radarsFanout); new_radar != nil {
			d.radars[if_name] = new_radar
			go d.radars[if_name].Run()
		}
		d.RadarsAlive = true
	}
	d.Unlock()
}

func (d *Discovery) radarRunner() {
	if d.RadarsAlive {
		return
	}
	for {
		for _, if_name := range d.interfaces {
			if if_name[len(if_name)-1] == '*' {
				//todo: handle an interface wildcard
			} else {
				d.AddRadar(if_name)
			}
		}
		// handle signals from Liana core
		select {
		case <-time.After(15 * time.Second):
			d.cfg.Log.Debug("Radar RUNNER timeout")
		case <-d.stopRadars: // stop all radars
			for _, r := range d.responders {
				r.stopResponder <- struct{}{}
			}
			for _, r := range d.radars {
				r.stopRadar <- struct{}{}
			}
			break
		}
	}
}

// /////

func (d *Discovery) Run() {
	if d.radarsFanout == nil {
		d.radarsFanout = make(chan string, RADARS_FANOUT_BUFFER)
	}
	if d.respondersFanout == nil {
		d.respondersFanout = make(chan string, RADARS_FANOUT_BUFFER)
	}
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
	return host_discovery
}

func init() {
	host_discovery = new(Discovery)
	host_discovery.radars = make(RadarMap)
	host_discovery.responders = make(ResponderMap)
}
