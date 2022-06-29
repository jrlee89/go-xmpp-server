package main

import (
	"crypto/tls"
    "github.com/jrlee89/go-xmpp-server/xmpp"
)

func main() {
	cert, _ := tls.LoadX509KeyPair("./lo.crt", "./lo.key")
    s := xmpp.NewServer(cert)
	s.Serve()
}
