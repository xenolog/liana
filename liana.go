package main

import (
	"fmt"
	// kcp "github.com/xtaci/kcp-go"
	// smux "github.com/xtaci/smux"
	cli "gopkg.in/urfave/cli.v1"
	"gopkg.in/xenolog/go-tiny-logger.v1"
	// "net"
	"os"
	// "sort"
	"github.com/xenolog/liana/config"
	"github.com/xenolog/liana/discovery"
	"github.com/xenolog/liana/identity"
	"strconv"
	"strings"
	"time"
)

const (
	Version = "0.0.1"
	CMD_LEN = 3 // do not decrease buffer size while redesign
)

var (
	Log *logger.Logger
	App *cli.App
	err error
)

func init() {
	// Setup logger
	Log = logger.New()

	// Configure CLI flags and commands
	App = cli.NewApp()
	App.Name = "RPC calls testing"
	App.Version = Version
	App.EnableBashCompletion = true
	// App.Usage = "Specify entry point of tree and got subtree for simple displaying"
	App.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Enable debug mode. Show more output",
		}, cli.BoolTFlag{
			Name:  "ipv4only",
			Usage: "Use only IPv4 addresses",
		}, cli.StringFlag{
			Name:  "url, u",
			Value: "udp://127.0.0.1:4001",
			Usage: "Specify URL for connection or listen",
		},
	}
	App.Commands = []cli.Command{{
		Name:   "server",
		Usage:  "run server",
		Action: runServer,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "password",
				Usage: "Specify password for crypt control traffic",
			}, cli.StringFlag{
				Name:  "interfaces",
				Usage: "Specify interfaces, where autodiscovering will be processed, use 'eth1,eth2' format",
			}, cli.StringFlag{
				Name:  "mcast-discovery",
				Value: "224.0.0.127:3232",
				Usage: "Specify mcast ipaddr, and port ",
			}, cli.UintFlag{
				Name:  "discovery-interval",
				Value: 20,
				Usage: "Specify mcast ipaddr, and port ",
			},
		},
	}, {
		Name:   "client",
		Usage:  "connect to server, ask to run simple job",
		Action: runClient,
	}}
	App.Before = func(c *cli.Context) error {
		if c.GlobalBool("debug") {
			Log.SetMinimalFacility(logger.LOG_D)
		} else {
			Log.SetMinimalFacility(logger.LOG_I)
		}
		Log.Debug("Started.")
		return nil
	}
	App.CommandNotFound = func(c *cli.Context, cmd string) {
		Log.Printf("Wrong command '%s'", cmd)
		os.Exit(1)
	}
}

func main() {
	App.Run(os.Args)
}

func Responder() {
	return
}

///// Server part /////

func runServer(c *cli.Context) error {
	password := c.String("password")
	if password == "" {
		Log.Error("Crypto key not defined.")
		os.Exit(1)
	}
	cfg := config.New()
	cfg.Log = Log
	cfg.McastInterval = time.Duration(c.Uint("discovery-interval"))
	cfg.Identity = identity.New(Log)
	cfg.McastDestination = c.String("mcast-discovery")
	cfg.ListenPort, _ = strconv.Atoi(strings.Split(c.String("mcast-discovery"), ":")[1])
	cfg.Identity.Run()
	Log.Info("Starting server (hostname=%s)", cfg.Identity.GetHostname())
	interfaces := strings.Split(c.String("interfaces"), ",")
	Log.Debug("Interfaces for autodiscovering: %s", interfaces)
	for _, iface := range interfaces {
		Log.Debug("+ starting discovery for '%s'", iface)
		r := discovery.New(cfg)
		if c.GlobalBool("ipv4only") {
			cfg.IPv4only = true
		}
		// go r.Run(iface, password)
		go r.Run()
	}
	time.Sleep(3 * time.Second)
	return nil
}

///// Client part /////

func runClient(c *cli.Context) error {
	fmt.Printf("Nothing to do...\n\n")
	return nil
}
