package composer

import (
	"google.golang.org/grpc/resolver"
)

const scheme = "transcoder"

func register(addrs []string) {
	resolverAddrs := make([]resolver.Address, 0, len(addrs))
	for _, addr := range addrs {
		resolverAddrs = append(resolverAddrs, resolver.Address{
			Addr: addr,
		})
	}
	resolver.Register(&resolverBuilder{
		addrs: resolverAddrs,
	})
}

type resolverBuilder struct {
	addrs []resolver.Address
}

func (b *resolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	cc.UpdateState(resolver.State{
		Addresses: b.addrs,
	})
	return &res{}, nil
}
func (*resolverBuilder) Scheme() string { return scheme }

type res struct{}

func (*res) ResolveNow(_ resolver.ResolveNowOptions) {}
func (*res) Close()                                  {}
