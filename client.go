package main

import (
	"crypto/tls"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
)

var debug io.Writer = os.Stdout

type client struct {
	conn          net.Conn
	p             *xml.Decoder
	e             *xml.Encoder
	jid           string
	msg           interface{}
	secure        bool
	authenticated bool
}

func (c *client) sendFeatures() {
	if !c.secure {
		c.restart()
		fmt.Fprintf(
			c.conn,
			"<stream:features><starttls xmlns='%s'><required/></starttls></stream:features>\n",
			nsTLS,
		)
		return
	}

	if !c.authenticated {
		c.restart()
		fmt.Fprintf(
			c.conn,
			"<stream:features><mechanisms xmlns='%s'><mechanism>ANONYMOUS</mechanism></mechanisms></stream:features>\n",
			nsSASL,
		)
		return
	}
	c.restart()
	fmt.Fprintf(
		c.conn,
		"<stream:features><bind xmlns='%s'/></stream:features>\n",
		nsBind,
	)
}

func (c *client) starttls(tc *tls.Config) error {
	fmt.Fprintf(c.conn, "<proceed xmlns='%s'/>\n", nsTLS)
	conn := tls.Server(c.conn, tc)
	err := conn.Handshake()
	if err != nil {
		fmt.Fprintf(c.conn, "<failure xmlns='%s'/>\n", nsTLS)
		return errors.New("starttls failure")
	}
	c.conn = conn
	c.p = xml.NewDecoder(c.conn)
	c.e = xml.NewEncoder(tee{c.conn, debug})
	c.secure = true
	return nil
}

func (c *client) auth(se xml.StartElement) error {
	for _, a := range se.Attr {
		switch a.Value {
		case "ANONYMOUS":
			fmt.Fprintf(c.conn, "<success xmlns='%s'/>", nsSASL)
			c.authenticated = true
			return nil
		}
	}
	fmt.Fprintf(
		c.conn,
		"<failure xmlns='%s'><malformed-request/></failure>",
		nsSASL,
	)
	return errors.New("auth failure")
}

func (c *client) restart() {
	fmt.Fprintf(
		c.conn,
		"<?xml version='1.0'?><stream:stream id='%x' version='1.0' xml:lang='en' xmlns:stream='%s' from='localhost' xmlns='%s'>\n",
		rng(),
		nsStreams,
		nsClient,
	)
}

func (c *client) bind(se xml.StartElement) error {
	var i iq
	//if err := c.p.DecodeElement(&i, nil); err != nil {
	if err := c.p.DecodeElement(&i, &se); err != nil {
		return errors.New("unmarshal <iq>: " + err.Error())
	}
	if &i.Bind == nil {
		fmt.Fprintf(
			c.conn,
			"<stream:error><not-well-formed xmlns='%s'/></stream:error>\n",
			nsStreams,
		)
		return errors.New("<iq> result missing <bind>")
	}
	c.jid = fmt.Sprintf("%x@localhost/%x", rng(), rng())
	fmt.Fprintf(
		c.conn,
		"<iq type='result' id='%x'><bind xmlns='%s'><jid>%s</jid></bind></iq>\n",
		&i.ID,
		nsBind,
		c.jid,
	)
	return nil
}

func nextStart(p *xml.Decoder) (xml.StartElement, error) {
	for {
		t, err := p.Token()
		if err != nil || t == nil {
			return xml.StartElement{}, err
		}
		switch t := t.(type) {
		case xml.StartElement:
			return t, nil
		}
	}
}
