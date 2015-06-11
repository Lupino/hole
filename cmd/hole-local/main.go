package main

import (
    "flag"
    "github.com/Lupino/hole"
)

var serverAddr string
var realAddr string
var certFile string
var privFile string
var useTLS bool

func init() {
    flag.StringVar(&serverAddr, "addr", "tcp://127.0.0.1:4000", "Hole server address.")
    flag.StringVar(&realAddr, "src", "tcp://127.0.0.1:8080", "Source server address.")
    flag.StringVar(&certFile, "cert", "cert.pem", "The cert file.")
    flag.StringVar(&privFile, "key", "cert.key", "The cert key file.")
    flag.BoolVar(&useTLS, "use-tls", false, "use TLS")
    flag.Parse()
}

func main() {
    var client = hole.NewClient(realAddr)
    if useTLS {
        client.ConfigTLS(certFile, privFile)
    }
    client.Connect(serverAddr)
    client.Process()
}
