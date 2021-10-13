// Copyright 2021 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package disco

import (
	"crypto"
	"encoding/xml"
)

// Caps can be included in a presence stanza or in stream features to advertise
// entity capabilities.
// Node is a string that uniquely identifies your client (eg.
// https://example.com/myclient) and ver is the hash of an Info value.
type Caps struct {
	XMLName xml.Name    `xml:"http://jabber.org/protocol/caps c"`
	Hash    crypto.Hash `xml:"hash,attr"`
	Node    string      `xml:"node,attr"`
	Ver     string      `xml:"ver,attr"`
}
