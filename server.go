package hole

import (
    "log"
    "net"
    "sync"
    "bytes"
    "strings"
    "github.com/satori/go.uuid"
)

type Server struct {
    clientConn Conn
    clientAlive bool
    alive bool
    sessions map[string]Session
    sessionLocker *sync.RWMutex
}

func NewServer() *Server {
    var server = new(Server)
    server.alive = true
    server.sessions = make(map[string]Session)
    server.clientAlive = false
    server.sessionLocker = new(sync.RWMutex)
    return server
}

func (server *Server) Serve(addr string) {
    parts := strings.SplitN(addr, "://", 2)
    listen, err := net.Listen(parts[0], parts[1])
    if err != nil {
        log.Fatal(err)
    }
    defer listen.Close()
    log.Printf("Hole proxy server started on %s\n", addr)
    for server.alive {
        conn, err := listen.Accept()
        if err != nil {
            log.Fatal(err)
        }
        if server.clientAlive {
            go server.handleConnection(conn)
        } else {
            go server.handleClient(conn)
        }
    }
}

func (server *Server) handleConnection(conn net.Conn) {
    sessionId := uuid.NewV4().Bytes()
    session := NewSession(sessionId, server.clientConn)
    server.sessionLocker.Lock()
    server.sessions[string(sessionId)] = session
    server.sessionLocker.Unlock()
    go PipeThenClose(conn, session.w)
    PipeThenClose(session.r, conn)
    server.sessionLocker.Lock()
    delete(server.sessions, string(session.Id))
    server.sessionLocker.Unlock()
}

func (server *Server) handleClient(conn net.Conn) {
    server.clientConn = NewServerConn(conn)
    server.clientAlive = true
    defer server.clientConn.Close()
    var err error
    var payload []byte
    if _, err = server.clientConn.Receive(); err != nil {
        server.clientAlive = false
        return
    }
    var sessionId, data []byte
    var ok bool
    var session Session
    for server.alive {
        if payload, err = server.clientConn.Receive(); err != nil {
            break
        }
        sessionId, data = DecodePacket(payload)
        server.sessionLocker.Lock()
        session, ok = server.sessions[string(sessionId)]
        server.sessionLocker.Unlock()
        if !ok {
            continue
        }
        if bytes.Equal(data, []byte("EOF")) {
            session.r.FeedEOF()
        } else {
            session.r.FeedData(data)
        }
    }
    server.clientAlive = false
}
