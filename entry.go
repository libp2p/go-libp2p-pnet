package pnet

import (
	"io"

	ipnet "github.com/libp2p/go-libp2p-interface-pnet"
	mc "github.com/multiformats/go-multicodec"
	bmux "github.com/multiformats/go-multicodec/base/mux"
)

func NewProtector(input io.Reader) (ipnet.Protector, error) {
	input = mc.WrapTransformPathToHeader(input)
	_ = bmux.AllBasesMux()

	return nil, nil
}
