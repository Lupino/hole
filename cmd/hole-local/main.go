package main

import (
	"flag"
	"github.com/Lupino/hole"
	"log"
	"time"
)

var serverAddr string
var realAddr string
var certFile string
var privFile string
var useTLS bool
var defaultReTryTime = 1000
var reTryTimes = defaultReTryTime

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

	for {
		if err := client.Connect(serverAddr); err != nil {
			reTryTimes = reTryTimes - 1
			if reTryTimes == 0 {
				break
			}
			log.Printf("Retry after 2 second...")
			time.Sleep(2 * time.Second)
			continue
		}
		client.Process()
		reTryTimes = defaultReTryTime
	}
}
