package coremock

import (
	"context"
	"fmt"

	libp2p2 "github.com/ipfs/go-ipfs/core/node/libp2p"

	"github.com/ipfs/go-ipfs/commands"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/repo"

	"github.com/ipfs/go-datastore"
	syncds "github.com/ipfs/go-datastore/sync"
	config "github.com/ipfs/go-ipfs-config"
	ma "github.com/multiformats/go-multiaddr"

	"github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-core/host"
	peer "github.com/libp2p/go-libp2p-core/peer"
	pstore "github.com/libp2p/go-libp2p-core/peerstore"
	testutil "github.com/libp2p/go-libp2p-testing/net"

	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
)

// NewMockNode constructs an IpfsNode for use in tests.
func NewMockNode() (*core.IpfsNode, error) {
	ctx := context.Background()

	// effectively offline, only peer in its network
	return core.NewNode(ctx, &core.BuildCfg{
		Online: true,
		Host:   MockHostOption(mocknet.New(ctx)),
	})
}

func MockHostOption(mn mocknet.Mocknet) libp2p2.HostOption {
	return func(ctx context.Context, id peer.ID, ps pstore.Peerstore, _ ...libp2p.Option) (host.Host, error) {
		host, err := mn.AddPeerWithPeerstore(id, ps)
		if err != nil {
			return nil, err
		}
		// Pretend to listen on a real network address
		// Otherwise, our fancy new DHT won't actually _form_ because
		// none of the nodes will be publicly dialable.
		port := id[len(id)-4]
		ipB := id[len(id)-3]
		ipC := id[len(id)-2]
		ipD := id[len(id)-1]
		return host, host.Network().Listen(ma.StringCast(fmt.Sprintf(
			"/ip4/18.%d.%d.%d/tcp/%d", port, ipB, ipC, ipD,
		)))
	}
}

func MockCmdsCtx() (commands.Context, error) {
	// Generate Identity
	ident, err := testutil.RandIdentity()
	if err != nil {
		return commands.Context{}, err
	}
	p := ident.ID()

	conf := config.Config{
		Identity: config.Identity{
			PeerID: p.String(),
		},
	}

	r := &repo.Mock{
		D: syncds.MutexWrap(datastore.NewMapDatastore()),
		C: conf,
	}

	node, err := core.NewNode(context.Background(), &core.BuildCfg{
		Repo: r,
	})
	if err != nil {
		return commands.Context{}, err
	}

	return commands.Context{
		ConfigRoot: "/tmp/.mockipfsconfig",
		LoadConfig: func(path string) (*config.Config, error) {
			return &conf, nil
		},
		ConstructNode: func() (*core.IpfsNode, error) {
			return node, nil
		},
	}, nil
}
