package pnet

import (
	"bytes"

	ipnet "github.com/libp2p/go-libp2p-interface-pnet"
)

func NewProtector(key []byte) (ipnet.Protector, error) {
	reader := bytes.NewReader(key)

	psk, err := decodeV1PSKKey(reader)
	if err != nil {
		return nil, err
	}
	return &protector{psk}, nil
}
