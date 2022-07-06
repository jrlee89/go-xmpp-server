package main

import (
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"log"
	"net"
)

type server struct {
	connections []*client
	transmit    chan *client
	register    chan *client
	unregister  chan *client
	tc          *tls.Config
}

func (s *server) listen() {
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
		go s.serve(c)
	}

}

func (s *server) run() {
	for {
		select {
		case message := <-s.transmit:
			s.send(message)
		case conn := <-s.register:
			s.connections = append(s.connections, conn)
		case conn := <-s.unregister:
			s.removeConn(conn)
		}
	}
}

func (s *server) serve(conn net.Conn) {
	c := &client{conn: conn, p: xml.NewDecoder(conn)}

	defer func() {
		fmt.Fprintf(c.conn, "</stream:stream>\n")
		c.conn.Close()
		s.unregister <- c
	}()

	for {
		se, _ := nextStart(c.p)
		switch se.Name.Local {
		case "stream":
			c.sendFeatures()
			break
		case "starttls":
			c.starttls(s.tc)
			break
		case "auth":
			c.auth(se)
			break
		case "iq":
			c.bind(se)
			s.register <- c
			break
		case "presence":
			c.msg = &presence{}
			s.transmit <- c
			break
		case "message":
			c.msg = &message{}
			if err := c.p.DecodeElement(c.msg, &se); err != nil {
				log.Printf("stream error: %s", err.Error())
				return
			}
			s.transmit <- c
			break
		}
	}
}

func (s *server) send(c *client) {
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

func (s *server) removeConn(c *client) {
	var i int
	for i = range s.connections {
		if s.connections[i].conn == c.conn {
			break
		}
	}
	s.connections = append(s.connections[:i], s.connections[i+1:]...)
}
