package pnet

import (
	"io"

	ipnet "github.com/libp2p/go-libp2p-interface-pnet"
)

func NewProtector(input io.Reader) (ipnet.Protector, error) {
	psk, err := decodeV1PSKKey(input)
	if err != nil {
		return nil, err
	}
	return &protector{psk}, nil
}
