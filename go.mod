module github.com/libp2p/dht-tracer1

go 1.12

require (
	github.com/ipfs/go-cid v0.0.2
	github.com/ipfs/go-ds-leveldb v0.0.2
	github.com/ipfs/go-ipfs-config v0.0.6
	github.com/ipfs/go-ipns v0.0.1 // indirect
	github.com/ipfs/go-log v0.0.1
	github.com/libp2p/go-libp2p v0.2.0-0.20190628095754-ccf9943938b9
	github.com/libp2p/go-libp2p-core v0.2.1-0.20190802003216-436d707f7cd0
	github.com/libp2p/go-libp2p-host v0.1.0 // indirect
	github.com/libp2p/go-libp2p-kad-dht v0.1.2-0.20190707121649-7b6f65e00898
	github.com/libp2p/go-libp2p-quic-transport v0.1.1
	github.com/libp2p/go-libp2p-record v0.1.0
	github.com/multiformats/go-multiaddr v0.0.4
)

replace github.com/libp2p/dht-tracer1 => ../dht-tracer1

replace github.com/libp2p/go-libp2p-kad-dht => ../go-libp2p-kad-dht

replace github.com/ipfs/go-todocounter => ../go-todocounter

replace github.com/libp2p/go-libp2p-core => ../go-libp2p-core
