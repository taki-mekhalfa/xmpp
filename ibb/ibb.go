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
	"errors"
	"net"
	"sync"

	"mellium.im/xmlstream"
	"mellium.im/xmpp"
	"mellium.im/xmpp/internal/attr"
	"mellium.im/xmpp/jid"
	"mellium.im/xmpp/mux"
	"mellium.im/xmpp/stanza"
)

// NS is the XML namespace used by IBB. It is provided as a convenience.
const NS = `http://jabber.org/protocol/ibb`

// BlockSize is the default block size in bytes used if an IBB stream is opened
// with no block size set.
// Because IBB base64 encodes the underlying data, the actual data transfered
// per stanza will be roughly twice the blocksize.
const BlockSize = 4096

const (
	messageType = "message"
	iqType      = "iq"
)

// Handle returns an option that registers a Handler for IBB payloads.
func Handle(h *Handler) mux.Option {
	return func(m *mux.ServeMux) {
		mux.IQ(stanza.SetIQ, xml.Name{Local: "open", Space: NS}, h)(m)
		mux.IQ(stanza.SetIQ, xml.Name{Local: "close", Space: NS}, h)(m)
		mux.IQ(stanza.SetIQ, xml.Name{Local: "data", Space: NS}, h)(m)
		mux.Message("", xml.Name{Local: "data", Space: NS}, h)(m)
	}
}

// Handler is an xmpp.Handler that handles multiplexing of bidirectional IBB
// streams.
type Handler struct {
	mu       sync.Mutex
	streams  map[string]*Conn
	listener map[string]*listener
}

// Listen returns a listener that accepts IBB streams.
func (h *Handler) Listen(addr jid.JID) net.Listener {
	if h.listener == nil {
		h.listener = make(map[string]*listener)
	}
	// TODO: in "open" check if listener is nil (or listener conn is closed) and
	// reject the connection with an error if so.
	if listener, ok := h.listener[addr.Bare().String()]; ok {
		return listener
	}
	listener := &listener{
		conn: make(chan net.Conn),
		addr: addr.Bare(),
	}
	h.listener[addr.Bare().String()] = listener
	return listener
}

// HandleMessage implements mux.MessageHandler.
func (h *Handler) HandleMessage(msg stanza.Message, t xmlstream.TokenReadEncoder) error {
	tok, err := t.Token()
	if err != nil {
		return err
	}
	// TODO: do we need to check this? Iterate through until we find the right
	// payload? I forget how this works.
	start := tok.(xml.StartElement)

	_ = start
	panic("ibb: message data not yet implemented")
}

// HandleIQ implements mux.IQHandler.
func (h *Handler) HandleIQ(iq stanza.IQ, re xmlstream.TokenReadEncoder, start *xml.StartElement) error {
	switch start.Name.Local {
	case "open":
		listener, ok := h.listener[iq.To.Bare().String()]
		if !ok {
			// If we're not listening for connections at this address, return an
			// error.
			// XEP-0047 §2.1:
			//     If the responder supports IBB but does not wish to proceed with the
			//     session, it returns a <not-acceptable/> error.
			return sendError(iq, re, stanza.Error{
				Type:      stanza.Cancel,
				Condition: stanza.NotAcceptable,
			})
		}

		_, sid := attr.Get(start.Attr, "sid")
		// TODO: somehow we need to get the session on the handler, but I don't see
		// how that's possible in a sane way.
		conn, err := newConn(h, s, iq)
		if err != nil {
			return err
		}
		h.addStream(sid, conn)
		return conn, nil
		listener.conn <- conn
	case "close":
		// TODO: if we receive a close element, should we flush any outgoing writes
		// first and make sure the conn is closed?
		// TODO: also check if the stream existed or not and return an error if they
		// tried to close a stream we weren't handling.
		_, sid := attr.Get(start.Attr, "sid")
		return h.closeSID(iq, re, sid)
	case "data":
		d := xml.NewTokenDecoder(re)
		p := dataPayload{}
		err := d.DecodeElement(&p, start)
		if err != nil {
			return err
		}
		return h.handlePayload(iq, re, p)
	}

	// TODO: error handling:
	//   Stanza errors of type wait that might mean we can resume later
	//   Because the session ID is unknown, the recipient returns an <item-not-found/> error with a type of 'cancel'.
	//   Because the sequence number has already been used, the recipient returns an <unexpected-request/> error with a type of 'cancel'.
	//   Because the data is not formatted in accordance with Section 4 of RFC 4648, the recipient returns a <bad-request/> error with a type of 'cancel'.
	// TODO: count seq numbers and close if out of order

	panic("not yet implemented")
}

func sendError(stanzaStart interface{}, e xmlstream.TokenReadEncoder, errPayload stanza.Error) error {
	switch s := stanzaStart.(type) {
	case stanza.Message:
		s.To, s.From = s.From, s.To
		s.Type = stanza.ErrorMessage
		_, err := xmlstream.Copy(e, s.Wrap(errPayload.TokenReader()))
		return err
	case stanza.IQ:
		_, err := xmlstream.Copy(e, s.Error(errPayload))
		return err
	}
	return errors.New("ibb: unexpected stanza type")
}

func (h *Handler) closeSID(stanzaStart interface{}, e xmlstream.TokenReadEncoder, sid string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	conn, ok := h.streams[sid]
	if !ok {
		// XEP-0047 Example 10. Recipient does not know about the IBB session
		// https://xmpp.org/extensions/xep-0047.html#example-10
		return sendError(stanzaStart, e, stanza.Error{
			Type:      stanza.Cancel,
			Condition: stanza.ItemNotFound,
		})
	}
	return conn.closeWithLock(false)
}

func (h *Handler) handlePayload(stanzaStart interface{}, e xmlstream.TokenReadEncoder, p dataPayload) error {
	//Seq     uint16   `xml:"seq,attr"`
	//SID     string   `xml:"sid,attr"`
	//data    []byte   `xml:",chardata"`
	h.mu.Lock()
	defer h.mu.Unlock()

	conn, ok := h.streams[p.SID]
	if !ok {
		return sendError(stanzaStart, e, stanza.Error{
			Type:      stanza.Cancel,
			Condition: stanza.ItemNotFound,
		})
	}

	// TODO: the XEP suggests that we only do this if the sequence number has
	// already been used, and just close it if we get an unexpected sequence
	// number, but surely this should be an error too?
	if p.Seq != conn.seqIn {
		return sendError(stanzaStart, e, stanza.Error{
			Type:      stanza.Cancel,
			Condition: stanza.UnexpectedRequest,
		})
	}

	conn.seqIn++
	_, err := conn.pw.Write(p.Data)
	return err
}

// Open attempts to create a new IBB stream on the provided session using IQs as
// the carrier stanza.
func (h *Handler) Open(ctx context.Context, s *xmpp.Session, to jid.JID, blockSize uint16) (*Conn, error) {
	return h.open(ctx, iqType, s, to, blockSize)
}

// OpenMessage attempts to create a new IBB stream on the provided session using
// messages as the carrier stanza.
// Most users should call Open instead.
func (h *Handler) OpenMessage(ctx context.Context, s *xmpp.Session, to jid.JID, blockSize uint16) (*Conn, error) {
	return h.open(ctx, messageType, s, to, blockSize)
}

func (h *Handler) open(ctx context.Context, stanzaType string, s *xmpp.Session, to jid.JID, blockSize uint16) (*Conn, error) {
	sid := attr.RandomID()

	iq := openIQ{
		IQ: stanza.IQ{
			To: to,
		},
	}
	iq.Open.SID = sid
	iq.Open.Stanza = stanzaType
	iq.Open.BlockSize = blockSize

	resp, err := s.SendIQ(ctx, iq.TokenReader())
	if err != nil {
		return nil, err
	}
	// TODO: resp should never be nil, is this something about the test
	// ClientServer?
	if resp != nil {
		defer resp.Close()
	}

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

	if h.streams == nil {
		h.streams = make(map[string]*Conn)
	}
	h.streams[sid] = conn
}
