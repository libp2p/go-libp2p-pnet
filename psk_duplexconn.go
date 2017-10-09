package pnet

import tpt "github.com/libp2p/go-libp2p-transport"

type pskDuplexConn struct {
	tpt.DuplexConn

	pskRW *pskReadWriter
}

var _ tpt.DuplexConn = &pskDuplexConn{}

func newPSKDuplexConn(psk *[32]byte, in tpt.DuplexConn) (*pskDuplexConn, error) {
	pskRW, err := newPSKReadWriter(psk, in)
	if err != nil {
		return nil, err
	}
	return &pskDuplexConn{
		DuplexConn: in,
		pskRW:      pskRW,
	}, nil
}

func (c *pskDuplexConn) Read(b []byte) (int, error) {
	return c.pskRW.Read(b)
}

func (c *pskDuplexConn) Write(p []byte) (int, error) {
	return c.pskRW.Write(p)
}
