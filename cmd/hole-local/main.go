package main

import (
    "flag"
    "github.com/Lupino/hole"
)

var serverAddr string
var realAddr string

func init() {
    flag.StringVar(&serverAddr, "addr", "tcp://127.0.0.1:4000", "Hole server address.")
    flag.StringVar(&realAddr, "src", "tcp://127.0.0.1:8080", "Source server address.")
    flag.Parse()
}

func main() {
    var client = hole.NewClient(realAddr)
    client.Connect(serverAddr)
    client.Process()
}
