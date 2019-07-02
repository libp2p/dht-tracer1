package dhttests

import (
  "bytes"
  "context"
  "fmt"
  "io"
  "math/rand"

  levelds "github.com/ipfs/go-ds-leveldb"
  ipfsconfig "github.com/ipfs/go-ipfs-config"
  ipns "github.com/ipfs/go-ipns"
  libp2p "github.com/libp2p/go-libp2p"
  peer "github.com/libp2p/go-libp2p-core/peer"
  host "github.com/libp2p/go-libp2p-host"
  dht "github.com/libp2p/go-libp2p-kad-dht"
  record "github.com/libp2p/go-libp2p-record"
  ping "github.com/libp2p/go-libp2p/p2p/protocol/ping"
)

var BootstrapAddrsStr = []string{
  "/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",  // mars.i.ipfs.io
  "/ip4/104.236.179.241/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM", // pluto.i.ipfs.io
  "/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu", // saturn.i.ipfs.io
  "/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64",   // venus.i.ipfs.io
  "/ip4/178.62.158.247/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd",  // earth.i.ipfs.io
  // "/ip6/2604:a880:1:20::203:d001/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",  // pluto.i.ipfs.io
  // "/ip6/2400:6180:0:d0::151:6001/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",  // saturn.i.ipfs.io
  // "/ip6/2604:a880:800:10::4a:5001/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64", // venus.i.ipfs.io
  // "/ip6/2a03:b0c0:0:1010::23:1001/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd", // earth.i.ipfs.io
}

var BootstrapAddrs []peer.AddrInfo

func init() {
  ais, err := ipfsconfig.ParseBootstrapPeers(BootstrapAddrsStr)
  if err != nil {
    panic(err)
  }
  BootstrapAddrs = ais
}

type Node struct {
  Host host.Host
  DHT  *dht.IpfsDHT
}

func (n *Node) ID() peer.ID {
  return n.Host.ID()
}

func (n *Node) Peers() []peer.ID {
  return n.Host.Network().Peers()
}

func (n *Node) RoutingTable() io.Reader {
  buf := bytes.NewBuffer(nil)
  PrintLatencyTable(buf, n.Host)
  return buf
}

func Bootstrap(n *Node) error {
  ctx := context.Background()

  // grab two random bootstrap nodes
  idx := rand.Intn(len(BootstrapAddrs) - 1)
  ais := BootstrapAddrs[idx : idx+2]

  for _, ai := range ais {
    err := n.Host.Connect(ctx, ai)
    if err != nil {
      return err
    }
  }

  err := n.DHT.Bootstrap(ctx)
  if err != nil {
    return err
  }

  // go do 5 RTTs
  for _, p := range n.Peers() {
    go func(p peer.ID) {
      ctx, cancel := context.WithCancel(ctx)
      defer cancel()

      // do 5 RTTs before canceling
      res := ping.Ping(ctx, n.Host, p)
      for i := 0; i < 5; i++ {
        <-res
      }
    }(p)
  }

  return nil
}

func NewNode() (*Node, error) {
  ds, err := levelds.NewDatastore("", nil)
  if err != nil {
    return nil, err
  }

  opts := []libp2p.Option{
    libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
  }

  h, err := libp2p.New(context.Background(), opts...)
  if err != nil {
    return nil, err
  }

  d := dht.NewDHT(context.Background(), h, ds)
  if err != nil {
    return nil, err
  }

  d.Validator = record.NamespacedValidator{
    "pk":   record.PublicKeyValidator{},
    "ipns": ipns.Validator{KeyBook: h.Peerstore()},
  }

  // dont need to store it, only mount it
  _ = ping.NewPingService(h)

  n := &Node{h, d}
  err = Bootstrap(n)
  return n, err
}

func PrintLatencyTable(w io.Writer, h host.Host) {
  ps := h.Network().Peers()
  fmt.Fprintf(w, "%v connected to %d peers\n", h.ID(), len(ps))
  for i, p := range ps {
    latency := h.Peerstore().LatencyEWMA(p)
    fmt.Fprintf(w, "%d %v %v\n", i, p, latency)
  }
}
