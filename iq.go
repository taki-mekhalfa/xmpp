// Copyright 2016 Sam Whited.
// Use of this source code is governed by the BSD 2-clause license that can be
// found in the LICENSE file.

package xmpp

import (
	"encoding/xml"
	"errors"
	"strings"

	"mellium.im/xmpp/jid"
)

// IQ ("Information Query") is used as a general request response mechanism.
// IQ's are one-to-one, provide get and set semantics, and always require a
// response in the form of a result or an error.
type IQ struct {
	ID      string   `xml:"id,attr"`
	Inner   string   `xml:",innerxml"`
	To      *jid.JID `xml:"to,attr"`
	From    *jid.JID `xml:"from,attr"`
	Lang    string   `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
	Type    iqType   `xml:"type,attr"`
	XMLName xml.Name `xml:"iq"`
}

type iqType int

const (
	// A GetIQ is used to query another entity for information.
	GetIQ iqType = iota

	// A SetIQ is used to provide data to another entity, set new values, and
	// replace existing values.
	SetIQ

	// A ResultIQ is sent in response to a successful get or set IQ.
	ResultIQ

	// An ErrorIQ is sent to report that an error occured during the delivery or
	// processing of a get or set IQ.
	ErrorIQ
)

func (t iqType) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	s := t.String()
	return xml.Attr{Name: name, Value: strings.ToLower(s[:len(s)-2])}, nil
}

func (t *iqType) UnmarshalXMLAttr(attr xml.Attr) error {
	switch attr.Value {
	case "get":
		*t = GetIQ
	case "set":
		*t = SetIQ
	case "result":
		*t = ResultIQ
	case "error":
		*t = ErrorIQ
	default:
		// TODO: This should be a stanza error with the bad-request condition.
		return errors.New("bad-request")
	}
	return nil
}
