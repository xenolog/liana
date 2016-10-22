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
	"strings"
	// "time"
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
		},
		cli.StringFlag{
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
			},
			cli.StringFlag{
				Name:  "interfaces",
				Usage: "Specify interfaces, where autodiscovering will be processed, use 'eth1,eth2' format",
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

func Radar(iface string, passwd string) {
	return
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
	Log.Info("Starting server")
	interfaces := strings.Split(c.String("interfaces"), ",")
	Log.Debug("Interfaces for autodiscovering: %s", interfaces)
	for _, iface := range interfaces {
		Log.Debug("+ starting radar for '%s'", iface)
		go Radar(iface, password)
	}
	return nil
}

///// Client part /////

func runClient(c *cli.Context) error {
	fmt.Printf("Nothing to do...\n\n")
	return nil
}
