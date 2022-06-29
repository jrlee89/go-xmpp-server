package xmpp

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
)

type server struct {
	connections []*conn
	unregister  chan *conn
	tx          chan *conn
	register    chan *conn
	tc          *tls.Config
}

func NewServer(cert tls.Certificate) *server {
	return &server{
		unregister: make(chan *conn),
		tx:         make(chan *conn),
		register:   make(chan *conn),
		tc:         &tls.Config{Certificates: []tls.Certificate{cert}},
	}
}

func (s *server) Serve() {
	ln, err := net.Listen("tcp", ":5222")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	go s.run()

	for {
		c, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go s.handler(c)
	}

}

func (s *server) run() {
	for {
		select {
		case message := <-s.tx:
			s.send(message)
		case conn := <-s.register:
			s.connections = append(s.connections, conn)
		case conn := <-s.unregister:
			s.removeConn(conn)
		}
	}
}

func (s *server) handler(conn net.Conn) {
    c := newConn(conn)
	for {
		se, _ := nextStart(c.p)
		switch se.Name.Local {
		case "stream":
			if err := c.openStream(s.tc); err != nil {
				log.Printf("stream error: %s", err.Error())
				fmt.Fprintf(c.conn, "</stream:stream>\n")
				c.conn.Close()
				return
			}
			s.register <- c
			break
		case "presence":
			c.msg = &presence{}
			s.tx <- c
			break
		case "message":
			c.msg = &message{}
			if err := c.p.DecodeElement(c.msg, &se); err != nil {
				s.unregister <- c
				c.conn.Close()
				return
			}
			s.tx <- c
			break
		}
	}
}

func (s *server) send(c *conn) {
	switch t := c.msg.(type) {
	case *message:
		for i := range s.connections {
			if t.To == s.connections[i].jid {
				s.connections[i].e.Encode(c.msg)
			}
		}
		return
	case *presence:
		for i := range s.connections {
			t.From = c.jid
			t.To = s.connections[i].jid
			s.connections[i].e.Encode(c.msg)
		}
		for i := range s.connections {
			if s.connections[i].jid != c.jid {
				t.From = s.connections[i].jid
				t.To = c.jid
				c.e.Encode(c.msg)
			}
		}
		return
	}
}

func (s *server) removeConn(c *conn) {
	var i int
	for i = range s.connections {
		if s.connections[i].conn == c.conn {
			break
		}
	}
	s.connections = append(s.connections[:i], s.connections[i+1:]...)
}
