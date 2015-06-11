package main

import (
    "flag"
    "github.com/Lupino/hole"
)

var serverAddr string
var certFile string
var privFile string
var useTLS bool

func init() {
    flag.StringVar(&serverAddr, "addr", "tcp://127.0.0.1:4000", "server address.")
    flag.StringVar(&certFile, "ca", "ca.pem", "The ca file.")
    flag.StringVar(&privFile, "key", "ca.key", "The ca key file.")
    flag.BoolVar(&useTLS, "use-tls", false, "use TLS")
    flag.Parse()
}

func main() {
    var server = hole.NewServer()
    if useTLS {
        server.ConfigTLS(certFile, privFile)
    }
    server.Serve(serverAddr)
}
