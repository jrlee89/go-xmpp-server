package main

import (
	"crypto/tls"
    "github.com/jrlee89/go-xmpp-server/xmpp"
)

func main() {
	cert, _ := tls.LoadX509KeyPair("./lo.crt", "./lo.key")
	s := &Server{
		unregister: make(chan *Conn),
		tx:         make(chan *Conn),
		register:   make(chan *Conn),
		tc:         &tls.Config{Certificates: []tls.Certificate{cert}},
	}
	s.Run()
}
