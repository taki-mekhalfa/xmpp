// Copyright 2017 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package disco_test

import (
	"bytes"
	"encoding/xml"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"mellium.im/xmlstream"
	"mellium.im/xmpp/disco"
	"mellium.im/xmpp/mux"
	"mellium.im/xmpp/stanza"
)

var _ mux.IQHandler = (*disco.Registry)(nil)

type Feature struct {
	Var string `xml:"var,attr"`
}

type Ident struct {
	Cat  string `xml:"category,attr"`
	Type string `xml:"type,attr"`
	Name string `xml:"name,attr"`
}

type Query struct {
	XMLName  xml.Name  `xml:"http://jabber.org/protocol/disco#info query"`
	Feature  []Feature `xml:"feature"`
	Identity []Ident   `xml:"identity"`
}

var registryTests = [...]struct {
	r   *disco.Registry
	q   Query
	err error
}{
	0: {
		q: Query{
			Feature: []Feature{
				{Var: disco.NSInfo},
				{Var: disco.NSItems},
			},
		},
	},
	1: {
		r: disco.NewRegistry(),
		q: Query{
			Feature: []Feature{
				{Var: disco.NSInfo},
				{Var: disco.NSItems},
			},
		},
	},
	2: {
		r: disco.NewRegistry(disco.Feature(disco.NSInfo), disco.Feature("porticulus")),
		q: Query{
			Feature: []Feature{
				{Var: disco.NSInfo},
				{Var: disco.NSItems},
				{Var: "porticulus"},
			},
		},
	},
	3: {
		r: disco.NewRegistry(disco.Feature(disco.NSItems), disco.AdminAccount("my service", "en"), disco.Feature("porticulus")),
		q: Query{
			Identity: []Ident{
				{"account", "admin", "my service"},
			},
			Feature: []Feature{
				{Var: disco.NSInfo},
				{Var: disco.NSItems},
				{Var: "porticulus"},
			},
		},
	},
}

func TestDisco(t *testing.T) {
	for i, tc := range registryTests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			d := xml.NewDecoder(strings.NewReader(`<query xmlns='http://jabber.org/protocol/disco#info'/>`))
			tok, _ := d.Token()
			start := tok.(xml.StartElement)
			buf := new(bytes.Buffer)
			e := xml.NewEncoder(buf)

			err := tc.r.HandleIQ(stanza.IQ{}, struct {
				xml.TokenReader
				xmlstream.Encoder
			}{
				TokenReader: d,
				Encoder:     e,
			}, &start)
			if err != tc.err {
				t.Fatalf("Unexpected error: want=`%v', got=`%v'", tc.err, err)
			}
			if err := e.Flush(); err != nil {
				t.Fatalf("Unexpected error while flushing: `%v'", err)
			}

			q := Query{}
			err = xml.Unmarshal(buf.Bytes(), &q)
			if err != nil {
				t.Fatalf("Unexpected error: `%v'", err)
			}

			// Clear the name so we don't have to set it in the test cases.
			q.XMLName = xml.Name{}

			sort.Slice(q.Feature, func(i, j int) bool {
				return q.Feature[i].Var < q.Feature[j].Var
			})
			sort.Slice(tc.q.Feature, func(i, j int) bool {
				return tc.q.Feature[i].Var < tc.q.Feature[j].Var
			})
			if !reflect.DeepEqual(q.Feature, tc.q.Feature) {
				t.Errorf("Features list did not match: want=`%+v', got=`%+v'", tc.q, q)
			}

			sort.Slice(q.Identity, func(i, j int) bool {
				return q.Identity[i].Cat+q.Identity[i].Type+q.Identity[i].Name < q.Identity[j].Cat+q.Identity[j].Type+q.Identity[j].Name
			})
			sort.Slice(tc.q.Identity, func(i, j int) bool {
				return tc.q.Identity[i].Cat+tc.q.Identity[i].Type+tc.q.Identity[i].Name < tc.q.Identity[j].Cat+tc.q.Identity[j].Type+tc.q.Identity[j].Name
			})
			if !reflect.DeepEqual(q.Identity, tc.q.Identity) {
				t.Errorf("Identity list did not match: want=`%+v', got=`%+v'", tc.q, q)
			}
		})
	}
}
