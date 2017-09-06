package pnet

import (
	"crypto/rand"
	"io"
	"net"
	"time"

	tpt "github.com/libp2p/go-libp2p-transport"
	smux "github.com/libp2p/go-stream-muxer"
	ma "github.com/multiformats/go-multiaddr"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func newDuplexStream() (io.ReadWriter, io.ReadWriter) {
	type rw struct {
		io.Reader
		io.Writer
	}

	r1, w1 := io.Pipe()
	r2, w2 := io.Pipe()
	return rw{r1, w2}, rw{r2, w1}
}

type mockDuplexConn struct {
	conn io.ReadWriter
}

var _ tpt.DuplexConn = &mockDuplexConn{}

func (c *mockDuplexConn) Read(b []byte) (int, error)       { return c.conn.Read(b) }
func (c *mockDuplexConn) Write(p []byte) (int, error)      { return c.conn.Write(p) }
func (c *mockDuplexConn) Close() error                     { panic("not implemented") }
func (c *mockDuplexConn) LocalAddr() net.Addr              { panic("not implemented") }
func (c *mockDuplexConn) LocalMultiaddr() ma.Multiaddr     { panic("not implemented") }
func (c *mockDuplexConn) RemoteAddr() net.Addr             { panic("not implemented") }
func (c *mockDuplexConn) RemoteMultiaddr() ma.Multiaddr    { panic("not implemented") }
func (c *mockDuplexConn) SetDeadline(time.Time) error      { panic("not implemented") }
func (c *mockDuplexConn) SetReadDeadline(time.Time) error  { panic("not implemented") }
func (c *mockDuplexConn) SetWriteDeadline(time.Time) error { panic("not implemented") }
func (c *mockDuplexConn) Transport() tpt.Transport         { panic("not implemented") }

type mockMultiplexConn struct {
	streamToAccept *mockStream
	streamToOpen   *mockStream
}

var _ tpt.MultiplexConn = &mockMultiplexConn{}

func (c *mockMultiplexConn) AcceptStream() (smux.Stream, error) { return c.streamToAccept, nil }
func (c *mockMultiplexConn) OpenStream() (smux.Stream, error)   { return c.streamToOpen, nil }
func (c *mockMultiplexConn) Close() error                       { panic("not implemented") }
func (c *mockMultiplexConn) IsClosed() bool                     { panic("not implemented") }
func (c *mockMultiplexConn) LocalAddr() net.Addr                { panic("not implemented") }
func (c *mockMultiplexConn) LocalMultiaddr() ma.Multiaddr       { panic("not implemented") }
func (c *mockMultiplexConn) RemoteAddr() net.Addr               { panic("not implemented") }
func (c *mockMultiplexConn) RemoteMultiaddr() ma.Multiaddr      { panic("not implemented") }
func (c *mockMultiplexConn) Transport() tpt.Transport           { panic("not implemented") }

type mockStream struct {
	io.ReadWriter
}

var _ smux.Stream = &mockStream{}

// func (s *mockStream) Read(b []byte) (int, error)       { return s.dataToRead.Read(b) }
// func (s *mockStream) Write(b []byte) (int, error)      { return s.dataToWrite.Write(b) }
func (s *mockStream) Close() error                     { panic("not implemented") }
func (s *mockStream) Reset() error                     { panic("not implemented") }
func (s *mockStream) SetDeadline(time.Time) error      { panic("not implemented") }
func (s *mockStream) SetReadDeadline(time.Time) error  { panic("not implemented") }
func (s *mockStream) SetWriteDeadline(time.Time) error { panic("not implemented") }

var _ = Describe("PSK protected SingleConn", func() {
	var (
		prot *protector
	)

	BeforeEach(func() {
		psk, err := GenerateV1PSK()
		Expect(err).ToNot(HaveOccurred())
		p, err := NewProtector(psk)
		Expect(err).ToNot(HaveOccurred())
		prot = p.(*protector)
	})

	Context("DuplexConns", func() {
		var conn1, conn2 *pskDuplexConn

		BeforeEach(func() {
			rw1, rw2 := newDuplexStream()
			c1, err := prot.Protect(&mockDuplexConn{rw1})
			Expect(err).ToNot(HaveOccurred())
			conn1 = c1.(*pskDuplexConn)
			c2, err := prot.Protect(&mockDuplexConn{rw2})
			Expect(err).ToNot(HaveOccurred())
			conn2 = c2.(*pskDuplexConn)
		})

		It("reads and writes on a DuplexConn", func(done Done) {
			testDone := make(chan struct{})
			// the connection is not buffered, so run it in a separate go-routine
			go func() {
				defer GinkgoRecover()
				// write a message
				_, err := conn1.Write([]byte("foobar"))
				Expect(err).ToNot(HaveOccurred())
				// read a message
				b := make([]byte, 6)
				_, err = io.ReadFull(conn1, b)
				Expect(err).ToNot(HaveOccurred())
				Expect(b).To(Equal([]byte("raboof")))
				close(testDone)
			}()

			// read a message
			b := make([]byte, 6)
			_, err := io.ReadFull(conn2, b)
			Expect(err).ToNot(HaveOccurred())
			Expect(b).To(Equal([]byte("foobar")))
			// write a message
			_, err = conn2.Write([]byte("raboof"))
			Expect(err).ToNot(HaveOccurred())
			<-testDone
			close(done)
		})

		It("does fragmented reads", func(done Done) {
			message := make([]byte, 1000)
			rand.Read(message)
			go func() {
				defer GinkgoRecover()
				_, err := conn1.Write(message)
				Expect(err).ToNot(HaveOccurred())
			}()

			b := make([]byte, 100)
			for i := 0; i < 10; i++ {
				n, err := conn2.Read(b)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(100))
				Expect(b).To(Equal(message[i*100 : i*100+100]))
			}
			close(done)
		})
	})

	It("reads and writes on a MultiplexConn", func(done Done) {
		rw1, rw2 := newDuplexStream()
		c1, err := prot.Protect(&mockMultiplexConn{streamToOpen: &mockStream{rw1}})
		Expect(err).ToNot(HaveOccurred())
		conn1 := c1.(*pskMultiplexConn)
		c2, err := prot.Protect(&mockMultiplexConn{streamToAccept: &mockStream{rw2}})
		Expect(err).ToNot(HaveOccurred())
		conn2 := c2.(*pskMultiplexConn)

		str1, err := conn1.OpenStream()
		Expect(err).ToNot(HaveOccurred())
		str2, err := conn2.AcceptStream()
		Expect(err).ToNot(HaveOccurred())

		testDone := make(chan struct{})
		// the connection is not buffered, so run it in a separate go-routine
		go func() {
			defer GinkgoRecover()
			// write a message
			_, err = str1.Write([]byte("foobar"))
			Expect(err).ToNot(HaveOccurred())
			// read a message
			b := make([]byte, 6)
			_, err = io.ReadFull(str1, b)
			Expect(err).ToNot(HaveOccurred())
			Expect(b).To(Equal([]byte("raboof")))
			close(testDone)
		}()

		// read a message
		b := make([]byte, 6)
		_, err = io.ReadFull(str2, b)
		Expect(err).ToNot(HaveOccurred())
		Expect(b).To(Equal([]byte("foobar")))
		// write a message
		_, err = str2.Write([]byte("raboof"))
		Expect(err).ToNot(HaveOccurred())
		<-testDone
		close(done)
	})
})
