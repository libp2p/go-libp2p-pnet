package pnet

import (
	smux "github.com/jbenet/go-stream-muxer"
	tpt "github.com/libp2p/go-libp2p-transport"
)

type pskMultiStreamConn struct {
	tpt.MultiStreamConn
	psk *[32]byte
}

var _ tpt.MultiStreamConn = &pskMultiStreamConn{}

func newPSKMultiStreamConn(psk *[32]byte, in tpt.MultiStreamConn) (*pskMultiStreamConn, error) {
	return &pskMultiStreamConn{
		MultiStreamConn: in,
		psk:             psk,
	}, nil
}

func (c *pskMultiStreamConn) AcceptStream() (smux.Stream, error) {
	stream, err := c.MultiStreamConn.AcceptStream()
	if err != nil {
		return nil, err
	}
	return newPSKStream(c.psk, stream)
}

func (c *pskMultiStreamConn) OpenStream() (smux.Stream, error) {
	stream, err := c.MultiStreamConn.OpenStream()
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
