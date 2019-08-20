package dhtnode

import (
	libp2p "github.com/libp2p/go-libp2p"
	peer "github.com/libp2p/go-libp2p-core/peer"
	quic "github.com/libp2p/go-libp2p-quic-transport"
)

type NodeCfg struct {
	Bootstrap  []*peer.AddrInfo
	Libp2pOpts []libp2p.Option
}

func DefaultNodeCfg() NodeCfg {
	return NodeCfg{
		Bootstrap:  BootstrapAddrs,
		Libp2pOpts: Libp2pOptionsTCP(),
	}
}

func Libp2pOptionsQUIC() []libp2p.Option {
	return []libp2p.Option{
		libp2p.Transport(quic.NewTransport),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/udp/0/quic"),
		libp2p.DisableRelay(),
	}
}

func Libp2pOptionsTCP() []libp2p.Option {
	return []libp2p.Option{
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		libp2p.DisableRelay(),
	}
}
