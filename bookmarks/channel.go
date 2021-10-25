// Copyright 2021 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package bookmarks

import (
	"bytes"
	"encoding/xml"
	"strconv"

	"mellium.im/xmlstream"
	"mellium.im/xmpp/jid"
)

// Bookmark represents a single chat room with various properties.
type Bookmark struct {
	JID        jid.JID
	Autojoin   bool
	Name       string
	Nick       string
	Password   string
	Extensions []byte
}

// TokenReader satisfies the xmlstream.Marshaler interface.
func (b Bookmark) TokenReader() xml.TokenReader {
	var payloads []xml.TokenReader
	if b.Nick != "" {
		payloads = append(payloads, xmlstream.Wrap(
			xmlstream.Token(xml.CharData(b.Nick)),
			xml.StartElement{
				Name: xml.Name{Local: "nick"},
			},
		))
	}
	if b.Password != "" {
		payloads = append(payloads, xmlstream.Wrap(
			xmlstream.Token(xml.CharData(b.Password)),
			xml.StartElement{
				Name: xml.Name{Local: "password"},
			},
		))
	}
	if len(b.Extensions) > 0 {
		payloads = append(payloads, xml.NewDecoder(bytes.NewReader(b.Extensions)))
	}
	conferenceAttrs := []xml.Attr{{
		Name:  xml.Name{Local: "autojoin"},
		Value: strconv.FormatBool(b.Autojoin),
	}}
	if b.Name != "" {
		conferenceAttrs = append(conferenceAttrs, xml.Attr{
			Name:  xml.Name{Local: "name"},
			Value: b.Name,
		})
	}

	return xmlstream.Wrap(
		xmlstream.MultiReader(payloads...),
		xml.StartElement{
			Name: xml.Name{Local: "conference", Space: NS},
			Attr: conferenceAttrs,
		},
	)
}

// WriteXML satisfies the xmlstream.WriterTo interface.
// It is like MarshalXML except it writes tokens to w.
func (b Bookmark) WriteXML(w xmlstream.TokenWriter) (n int, err error) {
	return xmlstream.Copy(w, b.TokenReader())
}

// MarshalXML implements xml.Marshaler.
func (b Bookmark) MarshalXML(e *xml.Encoder, _ xml.StartElement) error {
	_, err := b.WriteXML(e)
	return err
}

// UnmarshalXML implements xml.Unmarshaler.
func (b *Bookmark) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	data := struct {
		XMLName    xml.Name `xml:"urn:xmpp:bookmarks:1 conference"`
		Name       string   `xml:"name,attr"`
		Autojoin   bool     `xml:"autojoin,attr"`
		Nick       string   `xml:"nick"`
		Password   string   `xml:"password"`
		Extensions struct {
			Val []byte `xml:",innerxml"`
		} `xml:"extensions"`
	}{}
	err := d.DecodeElement(&data, &start)
	if err != nil {
		return err
	}

	b.Autojoin = data.Autojoin
	b.Name = data.Name
	b.Nick = data.Nick
	b.Password = data.Password
	b.Extensions = data.Extensions.Val
	return nil
}
