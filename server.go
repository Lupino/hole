package hole

import (
    "log"
    "net"
    "strings"
    "github.com/satori/go.uuid"
)

type Server struct {
    clientConn Conn
    clientAlive bool
    alive bool
}

func (server Server) Serve(addr string) {
    parts := strings.SplitN(addr, "://", 2)
    listen, err := net.Listen(parts[0], parts[1])
    if err != nil {
        log.Fatal(err)
    }
    defer listen.Close()
    log.Printf("Hole proxy server started on %s\n", addr)
    for {
        if !server.alive {
            break
        }
        conn, err := listen.Accept()
        if err != nil {
            log.Fatal(err)
        }
        if server.clientAlive {
            go server.handleConnection(conn)
        } else {
            server.RegisterClient(conn)
        }
    }
}

func (server *Server) handleConnection(conn net.Conn) {
    sessionId := uuid.NewV4().Bytes()
    session := NewSession(sessionId, server.clientConn)
    go PipeThenClose(conn, session.w)
    PipeThenClose(session.r, conn)
}

func (server *Server) RegisterClient(conn net.Conn) {
    server.clientConn = NewServerConn(conn)
    server.clientAlive = true
    if _, err := server.clientConn.Receive(); err != nil {
        server.clientAlive = false
        server.clientConn.Close()
    }
}
