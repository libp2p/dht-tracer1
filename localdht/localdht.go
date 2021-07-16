package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	logging "github.com/ipfs/go-log"
	dhtnode "github.com/libp2p/dht-tracer1/lib/dhtnode"
	peer "github.com/libp2p/go-libp2p-core/peer"
)

var Usage = `SYNOPSIS
    localdht - make a network of dht nodes

USAGE
    localdht [<opts>...] [<query>]

OPTIONS
    -h, --help        show usage
    -n <int>          number of dht nodes to run (default: 100)
    --debug           enable debug logging
    --quic            use quic transport only (helps w/ fd limits)
    --bootstrap-file  write bootstrap addresses to this file

EXAMPLES
    # run 100 dht nodes
    localdht
`

type Opts struct {
	BootstrapFile string
	NumNodes      int
	Debug         bool
	Quic          bool
}

func parseOpts() (Opts, []string) {
	var o Opts
	flag.IntVar(&o.NumNodes, "n", 100, "number of dht nodes to run")
	flag.BoolVar(&o.Debug, "debug", false, "enable debug logging")
	flag.BoolVar(&o.Quic, "quic", false, "use quic only")
	flag.StringVar(&o.BootstrapFile, "bootstrap-file", "", "bootstrap file")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, Usage)
	}
	flag.Parse()
	args := flag.Args()
	return o, args
}

func writeOutBootstrap(net *dhtnode.Net, file string) error {
	var w io.Writer
	if len(file) > 0 {
		f, err := os.Create(file)
		if err != nil {
			return err
		}
		defer f.Close()
		defer fmt.Println("wrote bootstrap addresses to:", file)
		w = f
	} else {
		w = os.Stdout
	}

	printAddrs(w, net.BootstrapAddrs)
	return nil
}

func runDHTNet(opts Opts) error {
	cfg := nodeCfgWithOpts(opts)

	net, err := dhtnode.NewNet(opts.NumNodes, cfg)
	if err != nil {
		return err
	}

	err = writeOutBootstrap(net, opts.BootstrapFile)
	if err != nil {
		return err
	}

	net.Bootstrap()

	// wait for termination. periodically print stats.
	terminate := termSignalChan()
	for {
		dhtnode.PrintNodeStats(os.Stdout, net.Nodes)

		select {
		case <-time.After(time.Second * 10):
		case <-terminate:
			fmt.Println("exiting...")
			return nil
		}
	}
}

func nodeCfgWithOpts(opts Opts) dhtnode.NodeCfg {
	cfg := dhtnode.DefaultNodeCfg()
	cfg.Bootstrap = nil // dont use any bootstrap addrs here

	if opts.Quic {
		cfg.Libp2pOpts = dhtnode.Libp2pOptionsQUIC()
	}

	return cfg
}

func errMain(opts Opts, _ []string) error {
	if opts.Debug {
		logging.SetLogLevel("tracedhtnode", "debug")
	}

	return runDHTNet(opts)
}

func main() {
	opts, args := parseOpts()
	if err := errMain(opts, args); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(-1)
	}
}

func termSignalChan() chan os.Signal {
	// wait until we exit.
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	return sigc
}

func printAddrs(w io.Writer, ais []*peer.AddrInfo) {
	mas, err := dhtnode.AddrInfosToP2pAddrs(ais)
	if err != nil {
		fmt.Fprintln(w, "error:", err)
		return
	}

	for _, ma := range mas {
		fmt.Fprintln(w, ma)
	}
}
