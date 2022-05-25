// Deprecated: This package has moved into go-libp2p as a sub-package: github.com/libp2p/go-libp2p/p2p/net/pnet.
package pnet

import (
	"github.com/libp2p/go-libp2p/p2p/net/pnet"
	"net"

	ipnet "github.com/libp2p/go-libp2p-core/pnet"
)

// NewProtectedConn creates a new protected connection
// Deprecated: use github.com/libp2p/go-libp2p/p2p/net/pnet.NewProtectedConn instead.
func NewProtectedConn(psk ipnet.PSK, conn net.Conn) (net.Conn, error) {
	return pnet.NewProtectedConn(psk, conn)
}
