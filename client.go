package hole

import (
    "net"
    "sync"
    "bytes"
    "strings"
)

type Client struct {
    sessions map[string]Session
    sessionLocker *sync.RWMutex
    conn Conn
    subAddr string
    alive bool
}

func NewClient(subAddr string) *Client {
    var client = new(Client)
    client.subAddr = subAddr
    client.sessions = make(map[string]Session)
    client.sessionLocker = new(sync.RWMutex)
    return client
}

func (client *Client) Connect(addr string) {
    parts := strings.SplitN(addr, "://", 2)
    var conn, err = net.Dial(parts[0], parts[1])
    client.alive = true
    if err != nil {
        client.alive = false
        return
    }

    client.conn = NewClientConn(conn)

    if err = client.conn.Send([]byte("Connected")); err != nil {
        client.alive = false
        client.conn.Close()
        return
    }
}

func (client *Client) Process() {
    defer client.conn.Close()
    var err error
    var payload []byte
    var sessionId, data []byte
    var session Session
    var ok bool
    for client.alive {
        if payload, err = client.conn.Receive(); err != nil {
            break
        }
        sessionId, data = DecodePacket(payload)
        client.sessionLocker.Lock()
        session, ok = client.sessions[string(sessionId)]
        client.sessionLocker.Unlock()
        if !ok {
            session = client.NewSession(sessionId)
            go client.handleSession(session)
        }
        if bytes.Equal(data, []byte("EOF")) {
            session.r.FeedEOF()
        } else {
            session.r.FeedData(data)
        }
    }
}

func (client *Client) NewSession(sessionId []byte) Session {
    var session = NewSession(sessionId, client.conn)
    client.sessionLocker.Lock()
    client.sessions[string(sessionId)] = session
    client.sessionLocker.Unlock()
    return session
}

func (client *Client) handleSession(session Session) {
    parts := strings.SplitN(client.subAddr, "://", 2)
    var conn, err = net.Dial(parts[0], parts[1])
    if err != nil {
        client.sessionLocker.Lock()
        delete(client.sessions, string(session.Id))
        client.sessionLocker.Unlock()
        return
    }
    go PipeThenClose(conn, session.w)
    PipeThenClose(session.r, conn)
    client.sessionLocker.Lock()
    delete(client.sessions, string(session.Id))
    client.sessionLocker.Unlock()
}
