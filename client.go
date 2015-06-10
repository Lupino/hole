package hole

import (
    "net"
    "log"
    "sync"
    "bytes"
    "strings"
    "crypto/tls"
    "crypto/x509"
    "io/ioutil"
)

type Client struct {
    sessions map[string]Session
    sessionLocker *sync.RWMutex
    conn Conn
    subAddr string
    alive bool
    tlsConfig tls.Config
    tls bool
}

func NewClient(subAddr string) *Client {
    var client = new(Client)
    client.subAddr = subAddr
    client.sessions = make(map[string]Session)
    client.sessionLocker = new(sync.RWMutex)
    client.tls = false
    return client
}

func (client *Client) ConfigTLS(certFile, privFile string) {
    client.tls = true
    cert2_b, _ := ioutil.ReadFile(certFile)
    priv2_b, _ := ioutil.ReadFile(privFile)
    priv2, _ := x509.ParsePKCS1PrivateKey(priv2_b)

    cert := tls.Certificate{
        Certificate: [][]byte{ cert2_b  },
        PrivateKey: priv2,
    }

    client.tlsConfig = tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
}

func (client *Client) Connect(addr string) {
    log.Printf("Connect Hole server: %s\n", addr)
    parts := strings.SplitN(addr, "://", 2)
    var conn net.Conn
    var err error
    if client.tls {
        conn, err = tls.Dial(parts[0], parts[1], &client.tlsConfig)
    } else {
        conn, err = net.Dial(parts[0], parts[1])
    }
    if err != nil {
        log.Fatal("Is the hole server started?")
        return
    }

    client.alive = true
    client.conn = NewClientConn(conn)

    if err := client.conn.Send([]byte("Connected")); err != nil {
        client.alive = false
        client.conn.Close()
        return
    }
}

func (client *Client) Process() {
    log.Printf("Process: %s\n", client.subAddr)
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
    log.Printf("New Session: %x\n", sessionId)
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
        log.Printf("Session: %x leave.\n", session.Id)
        client.sessionLocker.Unlock()
        return
    }
    go PipeThenClose(conn, session.w)
    PipeThenClose(session.r, conn)
    client.sessionLocker.Lock()
    delete(client.sessions, string(session.Id))
    log.Printf("Session: %x leave.\n", session.Id)
    client.sessionLocker.Unlock()
}
