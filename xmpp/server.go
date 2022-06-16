package xmpp

import (
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"log"
	"net"
)

type Server struct {
	connections []*Conn
	unregister  chan *Conn
	tx          chan *Conn
	register    chan *Conn
	tc          *tls.Config
}

func (s *Server) channels() {
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

func (s *Server) handleConn(conn net.Conn) {
	c := &Conn{
		conn: conn,
		p:    xml.NewDecoder(conn),
	}
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

func (s *Server) send(c *Conn) {
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

func (s *Server) removeConn(c *Conn) {
	var i int
	for i = range s.connections {
		if s.connections[i].conn == c.conn {
			break
		}
	}
	s.connections = append(s.connections[:i], s.connections[i+1:]...)
}
