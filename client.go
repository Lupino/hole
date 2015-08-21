package hole

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"sync"
)

// Client define a client.
type Client struct {
	sessions  map[string]Session
	locker    *sync.RWMutex
	conn      Conn
	subAddr   string
	alive     bool
	tlsConfig tls.Config
	useTLS    bool
}

// NewClient create a new client.
func NewClient(subAddr string) *Client {
	var client = new(Client)
	client.subAddr = subAddr
	client.sessions = make(map[string]Session)
	client.locker = new(sync.RWMutex)
	client.useTLS = false
	return client
}

// ConfigTLS config the client.
func (client *Client) ConfigTLS(certFile, privFile string) {
	client.useTLS = true
	cert2B, _ := ioutil.ReadFile(certFile)
	priv2B, _ := ioutil.ReadFile(privFile)
	priv2, _ := x509.ParsePKCS1PrivateKey(priv2B)

	cert := tls.Certificate{
		Certificate: [][]byte{cert2B},
		PrivateKey:  priv2,
	}

	client.tlsConfig = tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
}

// Connect connent to hole server.
func (client *Client) Connect(addr string) (err error) {
	log.Printf("Connect Hole server: %s\n", addr)
	parts := strings.SplitN(addr, "://", 2)
	var conn net.Conn
	if client.useTLS {
		conn, err = tls.Dial(parts[0], parts[1], &client.tlsConfig)
	} else {
		conn, err = net.Dial(parts[0], parts[1])
	}
	if err != nil {
		log.Printf("Is the hole server started?\n")
		return
	}

	client.alive = true
	client.conn = NewClientConn(conn)

	if err = client.conn.Send([]byte("Connected")); err != nil {
		client.alive = false
		client.conn.Close()
		return
	}

	return
}

// Process the client.
func (client *Client) Process() {
	log.Printf("Process: %s\n", client.subAddr)
	defer client.conn.Close()
	var err error
	var payload []byte
	var sessionID, data []byte
	var session Session
	var ok bool
	for client.alive {
		if payload, err = client.conn.Receive(); err != nil {
			break
		}
		sessionID, data = DecodePacket(payload)
		client.locker.Lock()
		session, ok = client.sessions[string(sessionID)]
		client.locker.Unlock()
		if !ok {
			session = client.NewSession(sessionID)
			go client.handleSession(session)
		}
		if bytes.Equal(data, []byte("EOF")) {
			session.r.FeedEOF()
		} else {
			session.r.FeedData(data)
		}
	}
	for _, session := range client.sessions {
		session.r.FeedEOF()
	}
	client.alive = false
}

// NewSession create a session.
func (client *Client) NewSession(sessionID []byte) Session {
	log.Printf("New Session: %x\n", sessionID)
	var session = NewSession(sessionID, client.conn)
	client.locker.Lock()
	client.sessions[string(sessionID)] = session
	client.locker.Unlock()
	return session
}

func (client *Client) handleSession(session Session) {
	parts := strings.SplitN(client.subAddr, "://", 2)
	var conn, err = net.Dial(parts[0], parts[1])
	if err != nil {
		client.locker.Lock()
		delete(client.sessions, string(session.ID))
		log.Printf("Session: %x leave.\n", session.ID)
		client.locker.Unlock()
		return
	}
	go PipeThenClose(conn, session.w)
	PipeThenClose(session.r, conn)
	client.locker.Lock()
	delete(client.sessions, string(session.ID))
	log.Printf("Session: %x leave.\n", session.ID)
	client.locker.Unlock()
}
