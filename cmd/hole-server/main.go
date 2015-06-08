package main

import (
    "flag"
    "github.com/Lupino/hole"
)

var serverAddr string

func init() {
    flag.StringVar(&serverAddr, "addr", "tcp://127.0.0.1:4000", "server address.")
    flag.Parse()
}

func main() {
    var server = hole.NewServer()
    server.Serve(serverAddr)
}
