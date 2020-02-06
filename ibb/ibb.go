// Copyright 2020 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

// Package ibb implements data transfer with XEP-0047: In-Band Bytestreams.
//
// In-band bytestreams (IBB) are a bidirectional data transfer mechanism that
// can be used to send small files or transfer other low-bandwidth data.
// Because IBB uses base64 encoding to send the binary data, it is extremely
// inefficient and should only be used as a fallback or last resort.
// When sending large amounts of data, a more efficient mechanism such as Jingle
// File Transfer (XEP-0234) or SOCKS5 Bytestreams (XEP-0065) should be used if
// possible.
package ibb // import "mellium.im/xmpp/ibb"

import (
	"context"
	"encoding/xml"
	"sync"

	"mellium.im/xmlstream"
	"mellium.im/xmpp"
	"mellium.im/xmpp/internal/attr"
	"mellium.im/xmpp/jid"
	"mellium.im/xmpp/mux"
	"mellium.im/xmpp/stanza"
)

// NS is the XML namespace used by IBB, provided as a convenience.
const NS = `http://jabber.org/protocol/ibb`

// BlockSize is the default block size used if an IBB stream is opened with no
// block size set.
// Because IBB base64 encodes the underlying data, the actual data transfered
// per stanza will be roughly twice the blocksize.
const BlockSize = 1 << 11

const (
	messageType = "message"
	iqType      = "iq"
)

// Handle is an option that registers a handler for all the correct stanza
// types and payloads.
func Handle(h *Handler) mux.Option {
	data := xml.Name{Space: NS, Local: "data"}
	return func(m *mux.ServeMux) {
		mux.Message("", data, h)(m)
		mux.IQ(stanza.SetIQ, data, h)(m)
		mux.IQ(stanza.SetIQ, xml.Name{Space: NS, Local: "open"}, h)(m)
		mux.IQ(stanza.SetIQ, xml.Name{Space: NS, Local: "close"}, h)(m)
	}
}

// Handler is an xmpp.Handler that handles multiplexing of bidirectional IBB
// streams.
type Handler struct {
	mu      sync.Mutex
	streams map[string]*Conn
}

// HandleXMPP implements xmpp.Handler.
func (h *Handler) HandleMessage(msg stanza.Message, t xmlstream.TokenReadEncoder) error {
	d := xml.NewTokenDecoder(t)

	p := dataMessage{}
	err := d.Decode(&p)
	if err != nil {
		return err
	}
	return handlePayload(p.Data)

	// TODO: error handling:
	//   Stanza errors of type wait that might mean we can resume later
	//   Because the session ID is unknown, the recipient returns an <item-not-found/> error with a type of 'cancel'.
	//   Because the sequence number has already been used, the recipient returns an <unexpected-request/> error with a type of 'cancel'.
	//   Because the data is not formatted in accordance with Section 4 of RFC 4648, the recipient returns a <bad-request/> error with a type of 'cancel'.
	// TODO: count seq numbers and close if out of order
}

// HandleIQ implements mux.IQHandler.
func (h *Handler) HandleIQ(iq stanza.IQ, t xmlstream.TokenReadEncoder, start *xml.StartElement) error {
	switch start.Name.Local {
	case "open":
		// TODO: add some sort of net.Listener based API for receiving conns
		panic("not yet implemented")
	case "close":
		// TODO: if we receive a close element, should we flush any outgoing writes
		// first and make sure the conn is closed?
		_, sid := attr.Get(start.Attr, "sid")
		h.rmStream(sid)
	case "data":
		d := xml.NewTokenDecoder(t)
		p := dataPayload{}
		err := d.DecodeElement(&p, start)
		if err != nil {
			return err
		}
		return handlePayload(p)
	}

	// TODO: error handling:
	//   Stanza errors of type wait that might mean we can resume later
	//   Because the session ID is unknown, the recipient returns an <item-not-found/> error with a type of 'cancel'.
	//   Because the sequence number has already been used, the recipient returns an <unexpected-request/> error with a type of 'cancel'.
	//   Because the data is not formatted in accordance with Section 4 of RFC 4648, the recipient returns a <bad-request/> error with a type of 'cancel'.
	// TODO: count seq numbers and close if out of order

	panic("not yet implemented")
}

func handlePayload(p dataPayload) error {
	panic("not yet implemented")
}

// Open attempts to create a new IBB stream on the provided session using IQs as
// the carrier stanza.
func (h *Handler) Open(ctx context.Context, s *xmpp.Session, to jid.JID, blockSize uint16) (*Conn, error) {
	return open(h, ctx, iqType, s, to, blockSize)
}

// OpenMessage attempts to create a new IBB stream on the provided session using
// messages as the carrier stanza.
// Most users should call Open instead.
func (h *Handler) OpenMessage(ctx context.Context, s *xmpp.Session, to jid.JID, blockSize uint16) (*Conn, error) {
	return open(h, ctx, messageType, s, to, blockSize)
}

func open(h *Handler, ctx context.Context, stanzaType string, s *xmpp.Session, to jid.JID, blockSize uint16) (*Conn, error) {
	sid := attr.RandomID()

	iq := openIQ{}
	iq.IQ.To = to
	iq.Open.SID = sid
	iq.Open.Stanza = stanzaType
	iq.Open.BlockSize = blockSize

	resp, err := s.SendIQ(ctx, iq.TokenReader())
	if err != nil {
		return nil, err
	}
	/* #nosec */
	defer resp.Close()

	conn, err := newConn(h, s, iq), nil
	if err != nil {
		return nil, err
	}
	h.addStream(sid, conn)
	return conn, nil
}

func (h *Handler) addStream(sid string, conn *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.streams[sid] = conn
}

func (h *Handler) rmStream(sid string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.streams, sid)
}
