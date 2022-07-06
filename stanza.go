package main

import "encoding/xml"

type message struct {
	XMLName xml.Name `xml:"jabber:client message"`
	From    string   `xml:"from,attr"`
	ID      string   `xml:"id,attr"`
	To      string   `xml:"to,attr"`
	Type    string   `xml:"type,attr"`
	Subject string   `xml:"subject"`
	Body    string   `xml:"body"`
	Thread  string   `xml:"thread"`
}

type iq struct {
	XMLName xml.Name `xml:"jabber:client iq"`
	ID      string   `xml:"id,attr"`
	Type    string   `xml:"type,attr"`
	Bind    bindBind
}

type bindBind struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-bind bind"`
}

type presence struct {
	XMLName xml.Name `xml:"jabber:client presence"`
	From    string   `xml:"from,attr"`
	To      string   `xml:"to,attr"`
}
