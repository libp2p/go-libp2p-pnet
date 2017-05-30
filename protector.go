package pnet

import (
	"errors"
	"io"

	ipnet "github.com/libp2p/go-libp2p-interface-pnet"
	tpt "github.com/libp2p/go-libp2p-transport"
)

type protector struct {
	psk         *[32]byte
	fingerprint []byte
}

var _ ipnet.Protector = (*protector)(nil)

// NewProtector creates ipnet.Protector instance from a io.Reader stream
// that should include Multicodec encoded V1 PSK.
func NewProtector(input io.Reader) (ipnet.Protector, error) {
	psk, err := decodeV1PSK(input)
	if err != nil {
		return nil, err
	}
	return NewV1ProtectorFromBytes(psk)
}

// NewV1ProtectorFromBytes creates ipnet.Protector of the V1 version.
func NewV1ProtectorFromBytes(psk *[32]byte) (ipnet.Protector, error) {
	return &protector{
		psk:         psk,
		fingerprint: fingerprint(psk),
	}, nil
}

func (p protector) Protect(in tpt.Conn) (tpt.Conn, error) {
	switch c := in.(type) {
	case tpt.SingleStreamConn:
		return newPSKSingleStreamConn(p.psk, c)
	case tpt.MultiStreamConn:
		return newPSKMultiStreamConn(p.psk, c)
	default:
		return nil, errors.New("connection is neither SingleStreamConn nor MultiStreamConn")
	}
}

func (p protector) Fingerprint() []byte {
	return p.fingerprint
}
