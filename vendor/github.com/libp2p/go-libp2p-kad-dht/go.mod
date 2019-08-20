module github.com/libp2p/go-libp2p-kad-dht

require (
	github.com/Kubuxu/go-more-timers v0.0.2
	github.com/gogo/protobuf v1.2.1
	github.com/hashicorp/golang-lru v0.5.1
	github.com/ipfs/go-cid v0.0.2
	github.com/ipfs/go-datastore v0.0.5
	github.com/ipfs/go-ipfs-util v0.0.1
	github.com/ipfs/go-log v0.0.1
	github.com/ipfs/go-todocounter v0.0.1
	github.com/jbenet/goprocess v0.1.3
	github.com/libp2p/go-eventbus v0.0.2
	github.com/libp2p/go-libp2p v0.2.0-0.20190628095754-ccf9943938b9
	github.com/libp2p/go-libp2p-circuit v0.1.0
	github.com/libp2p/go-libp2p-core v0.2.1-0.20190815235124-d29813389b68
	github.com/libp2p/go-libp2p-kbucket v0.2.0
	github.com/libp2p/go-libp2p-peerstore v0.1.2-0.20190621130618-cfa9bb890c1a
	github.com/libp2p/go-libp2p-record v0.1.0
	github.com/libp2p/go-libp2p-routing v0.1.0
	github.com/libp2p/go-libp2p-swarm v0.1.0
	github.com/libp2p/go-libp2p-testing v0.0.4
	github.com/libp2p/go-msgio v0.0.4
	github.com/mr-tron/base58 v1.1.2
	github.com/multiformats/go-multiaddr v0.0.4
	github.com/multiformats/go-multiaddr-dns v0.0.2
	github.com/multiformats/go-multiaddr-net v0.0.1
	github.com/multiformats/go-multistream v0.1.0
	github.com/stretchr/testify v1.3.0
	github.com/whyrusleeping/base32 v0.0.0-20170828182744-c30ac30633cc
	go.opencensus.io v0.21.0
	golang.org/x/xerrors v0.0.0-20190513163551-3ee3066db522
)

replace github.com/libp2p/go-libp2p-core => ../go-libp2p-core
