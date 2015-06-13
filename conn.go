package hole

import (
	"bytes"
	"errors"
	"net"
)

var (
	MAGIC_NOT_MATCH = errors.New("Magic not match.")
	MAGIC_REQUEST   = []byte("\x00REQ")
	MAGIC_RESPONSE  = []byte("\x00RES")
)

type Conn struct {
	net.Conn
	RequestMagic  []byte
	ResponseMagic []byte
}

// Create a connection
func NewConn(conn net.Conn, reqMagic, resMagic []byte) Conn {
	return Conn{Conn: conn, RequestMagic: reqMagic, ResponseMagic: resMagic}
}

// Create a server connection
func NewServerConn(conn net.Conn) Conn {
	return NewConn(conn, MAGIC_REQUEST, MAGIC_RESPONSE)
}

// Create a client connection
func NewClientConn(conn net.Conn) Conn {
	return NewConn(conn, MAGIC_RESPONSE, MAGIC_REQUEST)
}

// Receive waits for a new message on conn, and receives its payload.
func (conn *Conn) Receive() (rdata []byte, rerr error) {

	// Read magic
	magic, err := conn.receive(4)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(magic, conn.RequestMagic) {
		return magic, MAGIC_NOT_MATCH
	}

	// Read header
	header, err := conn.receive(4)
	if err != nil {
		return nil, err
	}

	length := ParseHeader(header)

	rdata, rerr = conn.receive(length)

	return
}

func (conn *Conn) receive(length uint32) ([]byte, error) {
	rdata := make([]byte, length)
	nRead := uint32(0)
	for nRead < length {
		n, err := conn.Read(rdata[nRead:])
		if err != nil {
			return nil, err
		}
		nRead = nRead + uint32(n)
	}
	return rdata, nil
}

// Send a new message.
func (conn *Conn) Send(data []byte) error {
	header, err := MakeHeader(data)
	if err != nil {
		return err
	}

	if err := conn.write(conn.ResponseMagic); err != nil {
		return err
	}

	if err := conn.write(header); err != nil {
		return err
	}

	if err := conn.write(data); err != nil {
		return err
	}

	return nil
}

func (conn *Conn) write(data []byte) error {
	written := 0
	length := len(data)
	for written < length {
		wrote, err := conn.Write(data[written:])
		if err != nil {
			return err
		}
		written = written + wrote
	}
	return nil
}
