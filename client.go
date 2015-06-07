package hole

import (
    "net"
    "bytes"
    "strings"
)

type Client struct {
    sessions map[string]Session
    conn Conn
    subAddr string
    alive bool
}

func (client *Client) handle(conn net.Conn) {
    client.conn = NewClientConn(conn)
    client.alive = true
    defer client.conn.Close()
    var err error
    var payload []byte
    if _, err = client.conn.Receive(); err != nil {
        client.alive = false
        return
    }
    var sessionId, data []byte
    var session Session
    var ok bool
    for {
        if payload, err = client.conn.Receive(); err != nil {
            break
        }
        sessionId, data = DecodePacket(payload)
        session, ok = client.sessions[string(sessionId)]
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
    client.sessions[string(sessionId)] = session
    return session
}

func (client *Client) handleSession(session Session) {
    parts := strings.SplitN(client.subAddr, "://", 2)
    var conn, err = net.Dial(parts[0], parts[1])
    if err != nil {
        delete(client.sessions, string(session.Id))
        return
    }
    go PipeThenClose(conn, session.w)
    PipeThenClose(session.r, conn)
    delete(client.sessions, string(session.Id))
}
