package hole

import (
    "io"
    "bytes"
    "sync"
)

type ReadStream struct {
    buffer [][]byte
    bufferSize int
    eof error
    locker *sync.RWMutex
    waiter *sync.RWMutex
    waiting bool
}

type WriteStream struct {
    sessionId []byte
    conn Conn
}

func NewReadStream() *ReadStream {
    var rs = new(ReadStream)
    rs.locker = new(sync.RWMutex)
    rs.waiter = new(sync.RWMutex)
    rs.waiting = false
    return rs
}

func (r *ReadStream) FeedData (buf []byte) {
    r.locker.Lock()
    r.buffer = append(r.buffer, buf)
    r.bufferSize = r.bufferSize + len(buf)
    r.locker.Unlock()

    if r.waiting {
        r.waiting = false
        r.waiter.Unlock()
    }
}

func (r *ReadStream) FeedEOF () {
    r.eof = io.EOF
    if r.waiting {
        r.waiting = false
        r.waiter.Unlock()
    }
}

func (r *ReadStream) Read(buf []byte) (length int, err error) {
    nRead := len(buf)
    for {
        if r.bufferSize > 0 || r.eof != nil {
            break
        }
        r.waiting = true
        r.waiter.Lock()
    }

    if r.bufferSize == 0 {
        err = r.eof
        return
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
    err = w.conn.Send(EncodePacket(w.sessionId, data))
    n = len(data)
    return
}

func (w *WriteStream) Close() error {
    // return w.conn.Close()
    _, err := w.Write([]byte("EOF"))
    return err
}
