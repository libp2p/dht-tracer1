package main

import (
  "flag"
  "fmt"
  "os"
  "time"

  dhttracer "github.com/libp2p/dht-tracer1/lib"
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
    curl "http://localhost:8080/events" | dht & >eventlogs
    curl "http://localhost:8080/cmd?q=put-value+foo+bar"
`

type Opts struct {
  ServerAddr string
  KadAlpha   int
}

func parseOpts() (Opts, []string) {
  var o Opts
  flag.StringVar(&o.ServerAddr, "serve", "localhost:7000", "http address for ctrl server")
  flag.IntVar(&o.KadAlpha, "alpha", 10, "alpha value for kad-dht")
  flag.Usage = func() {
    fmt.Fprintf(os.Stderr, Usage)
  }
  flag.Parse()
  args := flag.Args()
  return o, args
}

func runTracerServer(addr string) error {
  t := dhttracer.NewTracer()
  fmt.Println("dht node starting...")
  if err := t.Start(); err != nil {
    return err
  }

  // pause for a bit to let the node bootstrap.
  // (no nice way to listen for an event yet)
  time.Sleep(time.Second * 5)
  fmt.Println("dht node routing table:")
  fmt.Println(t.Node.RoutingTable())

  s := dhttracer.NewHTTPServer(t, addr)
  fmt.Println("server listening at", s.Server.Addr)
  return s.ListenAndServe() // hangs till done
}

func errMain(opts Opts, _ []string) error {

  dht.AlphaValue = opts.KadAlpha
  fmt.Fprintln(os.Stderr, "set dht.AlphaValue to", dht.AlphaValue)

  return runTracerServer(opts.ServerAddr)
}

func main() {
  opts, args := parseOpts()
  if err := errMain(opts, args); err != nil {
    fmt.Fprintln(os.Stderr, "error:", err)
    os.Exit(-1)
  }
}
