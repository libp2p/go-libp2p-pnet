package pnet

import (
	"crypto/cipher"
	"crypto/rand"
	"io"

	salsa20 "github.com/davidlazar/go-crypto/salsa20"
	mpool "github.com/jbenet/go-msgio/mpool"
	ipnet "github.com/libp2p/go-libp2p-interface-pnet"
)

// we are using buffer pool as user needs their slice back
// so we can't do XOR cripter in place
var (
	bufPool = &mpool.ByteSlicePool

	errShortNonce  = ipnet.NewError("could not read full nonce")
	errInsecureNil = ipnet.NewError("insecure is nil")
	errPSKNil      = ipnet.NewError("pre-shread key is nil")
)

type pskReadWriter struct {
	rw io.ReadWriter

	psk      *[32]byte
	writeS20 cipher.Stream
	readS20  cipher.Stream
}

func newPSKReadWriter(psk *[32]byte, insecure io.ReadWriter) (*pskReadWriter, error) {
	if insecure == nil {
		return nil, errInsecureNil
	}
	if psk == nil {
		return nil, errPSKNil
	}
	return &pskReadWriter{
		rw:  insecure,
		psk: psk,
	}, nil
}

func (c *pskReadWriter) Read(out []byte) (int, error) {
	if c.readS20 == nil {
		nonce := make([]byte, 24)
		_, err := io.ReadFull(c.rw, nonce)
		if err != nil {
			return 0, errShortNonce
		}
		c.readS20 = salsa20.New(c.psk, nonce)
	}

	maxn := uint32(len(out))
	in := bufPool.Get(maxn).([]byte) // get buffer
	defer bufPool.Put(maxn, in)      // put the buffer back

	in = in[:maxn]          // truncate to required length
	n, err := c.rw.Read(in) // read to in
	if err != nil {
		return 0, err
	}

	c.readS20.XORKeyStream(out[:n], in[:n]) // decrypt to out buffer

	return n, nil
}

func (c *pskReadWriter) Write(in []byte) (int, error) {
	if c.writeS20 == nil {
		nonce := make([]byte, 24)
		_, err := rand.Read(nonce)
		if err != nil {
			return 0, err
		}
		_, err = c.rw.Write(nonce)
		if err != nil {
			return 0, err
		}

		c.writeS20 = salsa20.New(c.psk, nonce)
	}
	n := uint32(len(in))
	out := bufPool.Get(n).([]byte) // get buffer
	defer bufPool.Put(n, out)      // put the buffer back

	out = out[:n]                    // truncate to required length
	c.writeS20.XORKeyStream(out, in) // encrypt

	return c.rw.Write(out) // send
}
