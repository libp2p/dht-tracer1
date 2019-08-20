package dhtnode

import (
	"fmt"
	"io"
	"sync"

	peer "github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

type Net struct {
	Nodes          []*Node
	BootstrapAddrs []*peer.AddrInfo
}

func NewNet(numNodes int, cfg NodeCfg) (*Net, error) {
	net := &Net{}
	nch := make(chan *Node, numNodes)

	// make nodes
	var wg sync.WaitGroup
	for i := 0; i < numNodes; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			n, err := NewNode(cfg)
			if err != nil {
				log.Error("network failed to start node", err)
				return
			}
			log.Debugf("created dht node %d %v", i, n)
			nch <- n
		}(i)
	}

	go func() {
		wg.Wait()
		close(nch)
	}()

	// this may result in len(net.Nodes) less than numNodes,
	// because an error in NewNode() will not send a node here.
	// for this tool, better to have a smaller network than panic
	// on a nil member in the array
	i := 0
	for n := range nch {
		net.Nodes = append(net.Nodes, n)
		i++
		if i%10 == 0 {
			log.Warningf("%d/%d nodes created", i, numNodes)
		}
	}

	net.BootstrapAddrs = cfg.Bootstrap
	if len(net.BootstrapAddrs) < 1 {
		// if no addrs given, use first 10 as bootstrappers.
		// more than 1 is useful to create an unevenly connected network.
		numForBootstrap := 10
		net.BootstrapAddrs = GetAddrInfos(net.Nodes[:numForBootstrap])
	}

	return net, nil
}

func (net *Net) Bootstrap() {
	log.Debug("bootstrapping network - start")

	var wg sync.WaitGroup
	for _, n := range net.Nodes {
		wg.Add(1)
		go func(n *Node) {
			defer wg.Done()
			err := Bootstrap(n, net.BootstrapAddrs)
			if err != nil {
				log.Error("failed to bootstrap", n, err)
			} else {
				log.Debug("bootstrapped", n)
			}
		}(n)
	}

	wg.Wait()
	log.Debug("bootstrapping network - end")
}

func GetAddrInfos(nodes []*Node) []*peer.AddrInfo {
	nodes2 := make([]*peer.AddrInfo, len(nodes))
	for i, n := range nodes {
		nodes2[i] = n.AddrInfo()
	}
	return nodes2
}

func PrintNodeStats(w io.Writer, nodes []*Node) {
	conns := 0
	for i, n := range nodes {
		id := n.Host.ID()
		ps := len(n.Host.Network().Peers())
		cs := len(n.Host.Network().Conns())
		fmt.Fprintf(w, "%v %v %d peers %d conns\n", i, id, ps, cs)
		conns += cs
	}
	fmt.Fprintf(w, "%v nodes, %v conns\n", len(nodes), conns)
}

// todo: move into go-libp2p-core/peer
func AddrInfosToP2pAddrs(ais []*peer.AddrInfo) ([]ma.Multiaddr, error) {
	var mas []ma.Multiaddr

	for _, ai := range ais {
		a, err := peer.AddrInfoToP2pAddrs(ai)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		mas = append(mas, a...)
	}
	return mas, nil
}
