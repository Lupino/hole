package main

import (
    "github.com/Lupino/hole"
)

func main() {
    var server = hole.NewServer()
    server.Serve("tcp://127.0.0.1:4000")
}
