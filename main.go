package main

import (
	"crypto/tls"
)

func main() {
	cert, _ := tls.LoadX509KeyPair("./lo.crt", "./lo.key")
	s := &server{
		transmit:   make(chan *client),
		register:   make(chan *client),
		unregister: make(chan *client),
		tc:         &tls.Config{Certificates: []tls.Certificate{cert}},
	}
	s.listen()
}
