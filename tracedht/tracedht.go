package main

import (
  "flag"
  "fmt"
  "os"
  "strings"
  "time"

  logging "github.com/ipfs/go-log"
  dhttracer "github.com/libp2p/dht-tracer1/lib"
  dhtnode "github.com/libp2p/dht-tracer1/lib/dhtnode"
  peer "github.com/libp2p/go-libp2p-core/peer"
  dht "github.com/libp2p/go-libp2p-kad-dht"
)

var Usage = `SYNOPSIS
    tracedht - trace dht queries to the ipfs dht

USAGE
    tracedht [<opts>...] [<query>]

OPTIONS
    -h, --help           show usage
    --serve <addr>       run ctrl http server on <addr>
    --kad-alpha <int>    set kad-dht alpha value (default: 10)
    --bootstrap <addrs>  non-default bootstrap multiaddrs (newline delimited)
    --debug              enable debug logs
    --quic               use quic transport only (helps with fd limits)
    #todo -f, --logfile  file to store eventlogs in

QUERIES
    Please see the documentation for libp2p-kad-dht to find out
    what dht queries mean and do. This tool assumes extensive
    familiarity with libp2p-kad-dht.

    tracedht queries are libp2p-kad-dht queries, expressed in
    text, with the following formats:

      put-value <key> <value>
      get-value <key> <value>
      add-provider <cid>
      get-providers <cid>
      find-peer <peer-id>
      ping <peer-id>

    Queries can be run via the commandline, or via an api server
    that this tool runs.


EXAMPLES
    # run server at localhost:8080
    tracedht --serve ":8080"

    # run dht queries w/ alpha value of 15
    tracedht --kad-alpha 15

    # run a specific query, and then Exit
    tracedht find-peer

    # server example
    tracedht --serve :8080 &
    curl "http://localhost:8080/cmd?q=put-value+foo+bar"
    curl "http://localhost:8080/cmd?q=find-peer+<peer-id>"

    # save event logs
    tracedht --serve :8080 &
    curl "http://localhost:8080/events" | grep dht >eventlogs
    curl "http://localhost:8080/cmd?q=put-value+foo+bar"
`

type Opts struct {
  Debug          bool
  ServerAddr     string
  KadAlpha       int
  BootstrapStr   string
  BootstrapAddrs []*peer.AddrInfo
  Quic           bool
}

func parseOpts() (Opts, []string, error) {
  var o Opts
  flag.BoolVar(&o.Debug, "debug", false, "enable debug logging")
  flag.BoolVar(&o.Quic, "quic", false, "quic transport only")
  flag.StringVar(&o.BootstrapStr, "bootstrap", "", "non default bootstrap addrs to use")
  flag.StringVar(&o.ServerAddr, "serve", "localhost:7000", "http address for ctrl server")
  flag.IntVar(&o.KadAlpha, "alpha", 10, "alpha value for kad-dht")
  flag.Usage = func() {
    fmt.Fprintf(os.Stderr, Usage)
  }
  flag.Parse()
  args := flag.Args()

  // bootstrap addr args
  o.BootstrapAddrs = dhtnode.BootstrapAddrs
  if o.BootstrapStr != "" {
    addrs := strings.Split(o.BootstrapStr, "\n")
    ais, err := dhtnode.ParseStrAddrs(addrs)
    if err != nil {
      return o, args, err
    }
    o.BootstrapAddrs = ais
  }

  return o, args, nil
}

func setupTracer(cfg dhtnode.NodeCfg) (*dhttracer.Tracer, error) {
  t := dhttracer.NewTracer(cfg)
  fmt.Println("dht node starting...")
  if err := t.Start(); err != nil {
    return nil, err
  }

  // pause for a bit to let the node bootstrap.
  // (no nice way to listen for an event yet)
  time.Sleep(time.Second * 5)
  fmt.Println("dht node routing table:")
  fmt.Println(t.Node.RoutingTable())
  return t, nil
}

func runTracerServer(t *dhttracer.Tracer, addr string) error {
  s := dhttracer.NewHTTPServer(t, addr)
  fmt.Println("server listening at", s.Server.Addr)
  return s.ListenAndServe() // hangs till done
}

func nodeCfgWithOpts(opts Opts) dhtnode.NodeCfg {
  cfg := dhtnode.DefaultNodeCfg()
  cfg.Bootstrap = opts.BootstrapAddrs

  if opts.Quic {
    cfg.Libp2pOpts = dhtnode.Libp2pOptionsQUIC()
  }

  return cfg
}

func errMain() error {
  opts, _, err := parseOpts()
  if err != nil {
    return err
  }

  // setup debug logging
  if opts.Debug {
    fmt.Fprintln(os.Stderr, "debug logging on")
    logging.SetLogLevel("dht", "debug")
    logging.SetLogLevel("tracedhtnode", "debug")
  }

  // update alpha value
  dht.AlphaValue = opts.KadAlpha
  fmt.Fprintln(os.Stderr, "set dht.AlphaValue to", dht.AlphaValue)

  // nodecfg
  cfg := nodeCfgWithOpts(opts)

  // setup tracer
  t, err := setupTracer(cfg)
  if err != nil {
    return err
  }

  // run tracer server
  return runTracerServer(t, opts.ServerAddr)
}

func main() {
  if err := errMain(); err != nil {
    fmt.Fprintln(os.Stderr, "error:", err)
    os.Exit(-1)
  }
}
