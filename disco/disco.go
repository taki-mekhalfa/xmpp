// Copyright 2017 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

// Package disco implements XEP-0030: Service Discovery.
package disco // import "mellium.im/xmpp/disco"

//go:generate go run gen.go

import (
	"encoding/xml"

	"mellium.im/xmlstream"
	"mellium.im/xmpp/internal/ns"
	"mellium.im/xmpp/stanza"
)

// Namespaces used by this package.
const (
	NSInfo  = `http://jabber.org/protocol/disco#info`
	NSItems = `http://jabber.org/protocol/disco#items`
)

type identity struct {
	Category string
	Type     string
	XMLLang  string
}

// A Registry is used to register features supported by a server.
type Registry struct {
	identities map[identity]string
	features   map[string]struct{}
}

// NewRegistry creates a new feature registry with the provided identities and
// features.
// If multiple identities are specified, the name of the registry will be used
// for all of them.
func NewRegistry(options ...Option) *Registry {
	registry := &Registry{
		features: map[string]struct{}{
			NSInfo:  struct{}{},
			NSItems: struct{}{},
		},
		identities: make(map[identity]string),
	}
	for _, o := range options {
		o(registry)
	}
	return registry
}

// HandleIQ handles disco info requests.
func (r *Registry) HandleIQ(iq stanza.IQ, t xmlstream.TokenReadEncoder, start *xml.StartElement) error {
	// TODO: decode the query including the node attribute and XML namespace so
	// that we can return the right things.
	if err := xmlstream.Skip(t); err != nil {
		return err
	}

	resp := xml.StartElement{
		Name: xml.Name{Space: NSInfo, Local: "query"},
	}
	if err := t.EncodeToken(resp); err != nil {
		return err
	}

	for feature := range r.features {
		start := xml.StartElement{
			Name: xml.Name{Local: "feature"},
			Attr: []xml.Attr{
				{Name: xml.Name{Local: "var"}, Value: feature},
			},
		}
		if err := t.EncodeToken(start); err != nil {
			return err
		}
		if err := t.EncodeToken(start.End()); err != nil {
			return err
		}
	}
	for ident, name := range r.identities {
		start := xml.StartElement{
			Name: xml.Name{Local: "identity"},
			Attr: []xml.Attr{
				{Name: xml.Name{Local: "category"}, Value: ident.Category},
				{Name: xml.Name{Local: "type"}, Value: ident.Type},
				{Name: xml.Name{Local: "name"}, Value: name},
				{Name: xml.Name{Space: ns.XML, Local: "lang"}, Value: ident.XMLLang},
			},
		}
		if err := t.EncodeToken(start); err != nil {
			return err
		}
		if err := t.EncodeToken(start.End()); err != nil {
			return err
		}
	}

	return t.EncodeToken(resp.End())
}
