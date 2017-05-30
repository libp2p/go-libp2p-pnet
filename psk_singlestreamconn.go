package pnet

import tpt "github.com/libp2p/go-libp2p-transport"

type pskSingleStreamConn struct {
	tpt.SingleStreamConn

	pskRW *pskReadWriter
}

var _ tpt.SingleStreamConn = &pskSingleStreamConn{}

func newPSKSingleStreamConn(psk *[32]byte, in tpt.SingleStreamConn) (*pskSingleStreamConn, error) {
	pskRW, err := newPSKReadWriter(psk, in)
	if err != nil {
		return nil, err
	}
	return &pskSingleStreamConn{
		SingleStreamConn: in,
		pskRW:            pskRW,
	}, nil
}

func (c *pskSingleStreamConn) Read(b []byte) (int, error) {
	return c.pskRW.Read(b)
}

func (c *pskSingleStreamConn) Write(p []byte) (int, error) {
	return c.pskRW.Write(p)
}
