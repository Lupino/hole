package main

import (
    "github.com/Lupino/hole"
)

func main() {
    var client = hole.NewClient("tcp://127.0.0.1:8080")
    client.Connect("tcp://127.0.0.1:4000")
    client.Process()
}
