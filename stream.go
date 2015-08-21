package hole

import (
	"bytes"
	"io"
	"sync"
)

// ReadStream define the base read stream.
type ReadStream struct {
	buffer     [][]byte
	bufferSize int
	eof        error
	locker     *sync.RWMutex
	waiter     *sync.RWMutex
	waiting    bool
}

// WriteStream define the base write stream.
type WriteStream struct {
	sessionID []byte
	conn      Conn
}

// NewReadStream create a read strean
func NewReadStream() *ReadStream {
	var rs = new(ReadStream)
	rs.locker = new(sync.RWMutex)
	rs.waiter = new(sync.RWMutex)
	rs.waiting = false
	return rs
}

// FeedData feed buffer from a socket or other.
func (r *ReadStream) FeedData(buf []byte) {
	r.locker.Lock()
	defer r.locker.Unlock()
	r.buffer = append(r.buffer, buf)
	r.bufferSize = r.bufferSize + len(buf)

	if r.waiting {
		r.waiting = false
		r.waiter.Unlock()
	}
}

// FeedEOF feed eof when the stream is closed.
func (r *ReadStream) FeedEOF() {
	r.locker.Lock()
	defer r.locker.Unlock()
	r.eof = io.EOF
	if r.waiting {
		r.waiting = false
		r.waiter.Unlock()
	}
}

func (r *ReadStream) Read(buf []byte) (length int, err error) {
	nRead := len(buf)
	for {
		r.locker.Lock()
		if r.bufferSize > 0 || r.eof != nil {
			r.locker.Unlock()
			break
		}
		r.waiting = true
		r.locker.Unlock()
		r.waiter.Lock()
	}

	err = r.eof
	if r.bufferSize == 0 {
		r.eof = io.EOF
		return 0, io.EOF
	}

	r.locker.Lock()
	data := bytes.Join(r.buffer, nil)
	if r.bufferSize >= nRead {
		copy(buf[0:], data[:nRead])
		r.buffer = [][]byte{data[nRead:]}
		r.bufferSize = r.bufferSize - nRead
		length = nRead
	} else {
		copy(buf[0:len(data)], data)
		r.buffer = [][]byte{}
		length = r.bufferSize
		r.bufferSize = 0
	}
	r.locker.Unlock()
	return
}

func (w *WriteStream) Write(data []byte) (n int, err error) {
	err = w.conn.Send(EncodePacket(w.sessionID, data))
	n = len(data)
	return
}

// Close the write stream.
func (w *WriteStream) Close() error {
	// return w.conn.Close()
	_, err := w.Write([]byte("EOF"))
	return err
}
