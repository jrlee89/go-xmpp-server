package xmpp

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

var debug io.Writer = os.Stdout

type conn struct {
	conn   net.Conn
	p      *xml.Decoder
	e      *xml.Encoder
	jid    string
	msg    interface{}
	secure bool
	sasl   bool
}

func newConn(c net.Conn) *conn {
	return &conn{conn: c, p: xml.NewDecoder(c)}
}

func (c *conn) openStream(tc *tls.Config) error {
	c.sendFeatures()
	if err := c.starttls(tc); err != nil {
		return err
	}
	if err := c.restart(); err != nil {
		return err
	}
	c.sendFeatures()
	if err := c.auth(); err != nil {
		return err
	}
	if err := c.restart(); err != nil {
		return err
	}
	c.sendFeatures()
	if err := c.bind(); err != nil {
		return err
	}
	return nil
}

func (c *conn) sendFeatures() {
	if !c.secure {
		fmt.Fprintf(
			c.conn,
			"<?xml version='1.0'?><stream:stream id='%x' version='1.0' xml:lang='en' xmlns:stream='%s' from='localhost' xmlns='%s'>\n",
			rng(),
			nsStream,
			nsClient,
		)
		fmt.Fprintf(
			c.conn,
			"<stream:features><starttls xmlns='%s'><required/></starttls></stream:features>\n",
			nsTLS,
		)
		return
	}

	if !c.sasl {
		fmt.Fprintf(
			c.conn,
			"<stream:features><mechanisms xmlns='%s'><mechanism>ANONYMOUS</mechanism></mechanisms></stream:features>\n",
			nsSASL,
		)
		return
	}
	fmt.Fprintf(
		c.conn,
		"<stream:features><bind xmlns='%s'/></stream:features>\n",
		nsBind,
	)
}

func (c *conn) starttls(tc *tls.Config) error {
	se, _ := nextStart(c.p)
	if se.Name.Local != "starttls" {
		fmt.Fprintf(c.conn, "<failure xmlns='%s'/>\n", nsTLS)
		return errors.New("starttls failure")
	}
	fmt.Fprintf(c.conn, "<proceed xmlns='%s'/>\n", nsTLS)
	conn := tls.Server(c.conn, tc)
	conn.Handshake()
	c.conn = conn
	c.p = xml.NewDecoder(c.conn)
	c.e = xml.NewEncoder(tee{c.conn, debug})
	c.secure = true
	return nil
}

func (c *conn) auth() error {
	se, _ := nextStart(c.p)
	for _, a := range se.Attr {
		switch a.Value {
		case "ANONYMOUS":
			fmt.Fprintf(c.conn, "<success xmlns='%s'/>", nsSASL)
			c.sasl = true
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

func (c *conn) restart() error {
	se, _ := nextStart(c.p)
	if se.Name.Local != "stream" {
		fmt.Fprintf(
			c.conn,
			"<stream:error><not-well-formed xmlns='%s'/></stream:error>\n",
			nsStream,
		)
		return errors.New("restart failed")
	}
	fmt.Fprintf(
		c.conn,
		"<?xml version='1.0'?><stream:stream id='%x' version='1.0' xml:lang='en' xmlns:stream='%s' from='localhost' xmlns='%s'>\n",
		rng(),
		nsStream,
		nsClient,
	)
	return nil
}

func (c *conn) bind() error {
	var i iq
	if err := c.p.DecodeElement(&i, nil); err != nil {
		return errors.New("unmarshal <iq>: " + err.Error())
	}
	if &i.Bind == nil {
		fmt.Fprintf(
			c.conn,
			"<stream:error><not-well-formed xmlns='%s'/></stream:error>\n",
			nsStream,
		)
		return errors.New("<iq> result missing <bind>")
	}
	c.jid = fmt.Sprintf("%x@localhost/%x", rng(), rng())
	fmt.Fprintf(
		c.conn,
		"<iq type='result' id='%x'><bind xmlns='%s'><jid>%s</jid></bind></iq>",
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

func rng() uint64 {
	var buf [8]byte
	if _, err := rand.Reader.Read(buf[:]); err != nil {
		log.Panic("Failed to read random bytes: " + err.Error())
	}
	return binary.LittleEndian.Uint64(buf[:])
}
