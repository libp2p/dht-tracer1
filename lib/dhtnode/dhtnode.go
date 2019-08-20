package dhtnode

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"strings"

	levelds "github.com/ipfs/go-ds-leveldb"
	ipfsconfig "github.com/ipfs/go-ipfs-config"
	// ipns "github.com/ipfs/go-ipns"
	logging "github.com/ipfs/go-log"
	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-core/host"
	peer "github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	record "github.com/libp2p/go-libp2p-record"
	ping "github.com/libp2p/go-libp2p/p2p/protocol/ping"
)

var log = logging.Logger("tracedhtnode")

var BootstrapAddrsStr = []string{
	"/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",  // mars.i.ipfs.io
	"/ip4/104.236.179.241/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM", // pluto.i.ipfs.io
	"/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu", // saturn.i.ipfs.io
	"/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64",   // venus.i.ipfs.io
	"/ip4/178.62.158.247/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd",  // earth.i.ipfs.io
	//
	//"/ip6/2604:a880:1:20::203:d001/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",  // pluto.i.ipfs.io
	//"/ip6/2400:6180:0:d0::151:6001/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",  // saturn.i.ipfs.io
	//"/ip6/2604:a880:800:10::4a:5001/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64", // venus.i.ipfs.io
	//"/ip6/2a03:b0c0:0:1010::23:1001/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd", // earth.i.ipfs.io
}

var BootstrapAddrs []*peer.AddrInfo

func init() {
	ais, err := ParseStrAddrs(BootstrapAddrsStr)
	if err != nil {
		panic(err)
	}

	BootstrapAddrs = ais
}

func ParseStrAddrs(addrs []string) ([]*peer.AddrInfo, error) {
	// remove garbage
	var addrs2 []string
	for _, a := range addrs {
		if len(a) < 1 {
			continue
		}
		if len(strings.Split(a, "/")) < 4 {
			continue
		}
		addrs2 = append(addrs2, a)
	}

	ais, err := ipfsconfig.ParseBootstrapPeers(addrs)
	if err != nil {
		return nil, err
	}

	ais2 := make([]*peer.AddrInfo, len(ais))
	for i, ai := range ais {
		ais2[i] = &peer.AddrInfo{}
		*ais2[i] = ai
	}
	return ais2, nil
}

type Node struct {
	Host host.Host
	DHT  *dht.IpfsDHT
}

func (n *Node) String() string {
	return n.Host.ID().String()
}

func (n *Node) ID() peer.ID {
	return n.Host.ID()
}

func (n *Node) Peers() []peer.ID {
	return n.Host.Network().Peers()
}

func (n *Node) AddrInfo() *peer.AddrInfo {
	return host.InfoFromHost(n.Host)
}

func (n *Node) RoutingTable() io.Reader {
	buf := bytes.NewBuffer(nil)
	PrintLatencyTable(buf, n.Host)
	return buf
}

func Bootstrap(n *Node, bootstrap []*peer.AddrInfo) error {
	ctx := context.Background()

	nb := 5 // number to bootstrap to
	var ais []*peer.AddrInfo
	for _, i := range rand.Perm(len(bootstrap)) {
		ais = append(ais, bootstrap[i])
		if len(ais) >= nb {
			break
		}
	}

	log.Debug("bootstrapping", n.ID(), "to", ais)

	for _, ai := range ais {
		err := n.Host.Connect(ctx, *ai)
		if err != nil {
			log.Error("failed bootstrap:", n.ID(), ai.ID, err)
		} else {
			log.Debug("bootstrapped:", n.ID(), ai.ID)
		}
	}

	if len(n.Peers()) < 1 {
		return fmt.Errorf("failed to bootstrap %s to any peer", n.ID())
	}

	err := n.DHT.Bootstrap(ctx)
	if err != nil {
		fmt.Printf("err bootstrapping:\n%v", err)
		return err
	}

	// go do 5 RTTs
	PingPeers(n, 5)

	return nil
}

func PingPeers(n *Node, rtts int) {
	for _, p := range n.Peers() {
		go func(p peer.ID) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// do 5 RTTs before canceling
			res := ping.Ping(ctx, n.Host, p)
			for i := 0; i < rtts; i++ {
				<-res
			}
		}(p)
	}
}

func NewNode(cfg NodeCfg) (*Node, error) {
	ds, err := levelds.NewDatastore("", nil)
	if err != nil {
		fmt.Printf("err creating new datastore:\n%v", err)
		return nil, err
	}

	h, err := libp2p.New(context.Background(), cfg.Libp2pOpts...)
	if err != nil {
		fmt.Printf("err creating new libptp:\n%v", err)
		return nil, err
	}

	d := dht.NewDHT(context.Background(), h, ds)
	if err != nil {
		fmt.Printf("err creating new dht:\n%v", err)
		return nil, err
	}

	d.Validator = record.NamespacedValidator{
		"pk": record.PublicKeyValidator{},
		// "ipns": ipns.Validator{KeyBook: h.Peerstore()},
		"v": blankValidator{},
	}

	// dont need to store it, only mount it
	_ = ping.NewPingService(h)
	n := &Node{h, d}

	if len(cfg.Bootstrap) > 0 {
		err = Bootstrap(n, cfg.Bootstrap)
	}
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

type blankValidator struct{}

func (blankValidator) Validate(_ string, _ []byte) error        { return nil }
func (blankValidator) Select(_ string, _ [][]byte) (int, error) { return 0, nil }
