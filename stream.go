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
}

type WriteStream struct {
    sessionId []byte
    conn Conn
}

func (r *ReadStream) FeedData (buf []byte) {
    r.locker.Lock()
    r.buffer = append(r.buffer, buf)
    r.bufferSize = r.bufferSize + len(buf)
    r.locker.Unlock()

    r.waiter.Unlock()
}

func (r *ReadStream) FeedEOF () {
    r.eof = io.EOF
    r.waiter.Unlock()
}

func (r *ReadStream) Read(buf []byte) (length int, err error) {
    nRead := len(buf)
    for {
        if r.bufferSize >= nRead || r.eof != nil {
            break
        }
        r.waiter.Lock()
    }

    if r.bufferSize == 0 {
        err = r.eof
        return
    }

    r.locker.Lock()
    data := bytes.Join(r.buffer, nil)
    if r.bufferSize > nRead {
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
    n, err = w.conn.Write(EncodePacket(w.sessionId, data))
    return
}

func (w *WriteStream) Close() error {
    // return w.conn.Close()
    return nil
}
