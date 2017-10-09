package pnet

import (
	tpt "github.com/libp2p/go-libp2p-transport"
	smux "github.com/libp2p/go-stream-muxer"
)

type pskMultiplexConn struct {
	tpt.MultiplexConn
	psk *[32]byte
}

var _ tpt.MultiplexConn = &pskMultiplexConn{}

func newPSKMultiplexConn(psk *[32]byte, in tpt.MultiplexConn) (*pskMultiplexConn, error) {
	return &pskMultiplexConn{
		MultiplexConn: in,
		psk:           psk,
	}, nil
}

func (c *pskMultiplexConn) AcceptStream() (smux.Stream, error) {
	stream, err := c.MultiplexConn.AcceptStream()
	if err != nil {
		return nil, err
	}
	return newPSKStream(c.psk, stream)
}

func (c *pskMultiplexConn) OpenStream() (smux.Stream, error) {
	stream, err := c.MultiplexConn.OpenStream()
	if err != nil {
		return nil, err
	}
	return newPSKStream(c.psk, stream)
}

type pskStream struct {
	smux.Stream
	pskRW *pskReadWriter
}

var _ smux.Stream = &pskStream{}

func newPSKStream(psk *[32]byte, stream smux.Stream) (smux.Stream, error) {
	pskRW, err := newPSKReadWriter(psk, stream)
	if err != nil {
		return nil, err
	}
	return &pskStream{
		Stream: stream,
		pskRW:  pskRW,
	}, nil
}

func (s *pskStream) Read(b []byte) (int, error) {
	return s.pskRW.Read(b)
}

func (s *pskStream) Write(p []byte) (int, error) {
	return s.pskRW.Write(p)
}
