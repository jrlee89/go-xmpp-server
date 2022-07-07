package main

import (
	"crypto/tls"
	"log"
	"os"
)

func main() {
	cert, _ := tls.LoadX509KeyPair("./lo.crt", "./lo.key")
	logger := log.New(os.Stderr, "XMPP Error: ", log.LstdFlags)
	s := &server{
		hostname:   "localhost",
		transmit:   make(chan *client),
		register:   make(chan *client),
		unregister: make(chan *client),
		tc:         &tls.Config{Certificates: []tls.Certificate{cert}},
		msgLog:     os.Stdout,
		errLog:     logger,
	}
	s.listen()
}
