package pnet

import (
	iconn "github.com/libp2p/go-libp2p-interface-conn"
	ipnet "github.com/libp2p/go-libp2p-interface-pnet"
)

type protector struct {
	psk *[32]byte
}

var _ ipnet.Protector = (*protector)(nil)

func (p protector) Protect(in iconn.Conn) (iconn.Conn, error) {
	return newPSKConn(p.psk, in)
}
