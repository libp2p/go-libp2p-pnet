package pnet

import (
	"crypto/cipher"
	"crypto/rand"
	"errors"

	salsa20 "github.com/davidlazar/go-crypto/salsa20"
	mpool "github.com/jbenet/go-msgio/mpool"
	iconn "github.com/libp2p/go-libp2p-interface-conn"
)

// we are using buffer pool as user needs their slice back
// so we can't do XOR cripter in place
var bufPool = mpool.ByteSlicePool

type pskConn struct {
	iconn.Conn
	psk *[32]byte

	writeS20 cipher.Stream
	readS20  cipher.Stream
}

func (c *pskConn) Read(out []byte) (int, error) {
	if c.readS20 == nil {
		nonce := make([]byte, 24)
		n, err := c.Conn.Read(nonce)
		if err != nil {
			return 0, err
		}
		if n != 24 {
			return 0, errors.New("could not read full nonce")
		}
		c.readS20 = salsa20.New(c.psk, nonce)
	}

	maxn := uint32(len(out))
	in := bufPool.Get(maxn).([]byte) // get buffer
	defer bufPool.Put(maxn, in)      // put the buffer back

	in = in[:maxn]
	n, err := c.Conn.Read(in)
	if err != nil {
		return 0, err
	}

	c.readS20.XORKeyStream(out[:n], in[:n])

	return n, nil
}

func (c *pskConn) Write(in []byte) (int, error) {
	if c.writeS20 == nil {
		nonce := make([]byte, 24)
		_, err := rand.Read(nonce)
		if err != nil {
			return 0, err
		}
		_, err = c.Conn.Write(nonce)
		if err != nil {
			return 0, err
		}

		c.writeS20 = salsa20.New(c.psk, nonce)
	}
	n := uint32(len(in))
	out := bufPool.Get(n).([]byte) // get buffer
	defer bufPool.Put(n, out)      // put the buffer back

	out = out[:n]
	c.writeS20.XORKeyStream(out, in)

	return c.Conn.Write(out)
}

var _ iconn.Conn = (*pskConn)(nil)

func newPSKConn(psk *[32]byte, insecure iconn.Conn) (iconn.Conn, error) {
	if insecure == nil {
		return nil, errors.New("insecure is nil")
	}
	if psk == nil {
		return nil, errors.New("pre-shread key is nil")
	}
	return &pskConn{
		Conn: insecure,
		psk:  psk,
	}, nil
}
