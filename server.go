package hole

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"github.com/satori/go.uuid"
	"io"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"sync"
)

type Server struct {
	clientConn    Conn
	clientAlive   bool
	alive         bool
	sessions      map[string]Session
	sessionLocker *sync.RWMutex
	tlsConfig     tls.Config
	tls           bool
}

func NewServer() *Server {
	var server = new(Server)
	server.alive = true
	server.sessions = make(map[string]Session)
	server.clientAlive = false
	server.sessionLocker = new(sync.RWMutex)
	server.tls = false
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
		if server.clientIsAlive() {
			go server.handleConnection(conn)
		} else {
			go server.handleClient(conn)
		}
	}
}

func (server *Server) RegisterClient(conn net.Conn) {
	server.sessionLocker.Lock()
	defer server.sessionLocker.Unlock()

	server.clientConn = NewServerConn(conn)
	server.clientAlive = true
	var err error
	if _, err = server.clientConn.Receive(); err != nil {
		server.clientAlive = false
		return
	}
}

func (server *Server) clientIsAlive() bool {
	server.sessionLocker.Lock()
	defer server.sessionLocker.Unlock()

	return server.clientAlive
}

func (server *Server) ConfigTLS(certFile, privFile string) {
	ca_b, _ := ioutil.ReadFile(certFile)
	ca, _ := x509.ParseCertificate(ca_b)
	priv_b, _ := ioutil.ReadFile(privFile)
	priv, _ := x509.ParsePKCS1PrivateKey(priv_b)

	pool := x509.NewCertPool()
	pool.AddCert(ca)

	cert := tls.Certificate{
		Certificate: [][]byte{ca_b},
		PrivateKey:  priv,
	}

	server.tlsConfig = tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    pool,
	}
	server.tlsConfig.Rand = rand.Reader
	server.tls = true
}

func (server *Server) handleConnection(conn net.Conn) {
	log.Printf("Handle connection: %s\n", conn.RemoteAddr().String())
	server.sessionLocker.Lock()
	sessionId := uuid.NewV4().Bytes()
	session := NewSession(sessionId, server.clientConn)
	server.sessions[string(sessionId)] = session
	server.sessionLocker.Unlock()
	go PipeThenClose(conn, session.w)
	PipeThenClose(session.r, conn)
	server.sessionLocker.Lock()
	delete(server.sessions, string(session.Id))
	server.sessionLocker.Unlock()
}

func (server *Server) handleClient(conn net.Conn) {
	log.Printf("New Client: %s\n", conn.RemoteAddr().String())
	if server.tls {
		conn = tls.Server(conn, &server.tlsConfig)
	}
	server.RegisterClient(conn)
	defer server.clientConn.Close()
	var err error
	var payload []byte
	var sessionId, data []byte
	var ok bool
	var session Session
	for server.alive {
		if payload, err = server.clientConn.Receive(); err != nil {
			if err == io.EOF {
				log.Printf("Client: %s leave.\n", conn.RemoteAddr().String())
			} else {
				log.Printf("Error: %s\n", err.Error())
			}
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
	for _, session := range server.sessions {
		session.r.FeedEOF()
	}
	server.sessionLocker.Lock()
	server.clientAlive = false
	server.sessionLocker.Unlock()
}
