// Copyright 2021 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package muc

import (
	"context"
	"encoding/xml"

	"mellium.im/xmlstream"
	"mellium.im/xmpp"
	"mellium.im/xmpp/internal/attr"
	"mellium.im/xmpp/jid"
	"mellium.im/xmpp/stanza"
)

// Channel represents a group chat, conference, or chatroom.
//
// Channel aims to be as stateless as possible, so details such as the channel
// subject and participant list are not stored.
// Instead, it is up to the user to store this information and associate it with
// the channel (probably by mapping details to the channel address).
type Channel struct {
	addr    jid.JID
	pass    string
	client  *Client
	session *xmpp.Session

	join   chan jid.JID
	depart chan struct{}
}

// Addr returns the address of the channel.
func (c *Channel) Addr() jid.JID {
	return c.addr.Bare()
}

// Me returns the users last-known address in the channel.
func (c *Channel) Me() jid.JID {
	return c.addr
}

// Joined returns true if this room is still being managed by the service.
func (c *Channel) Joined() bool {
	c.client.managedM.Lock()
	defer c.client.managedM.Unlock()
	_, ok := c.client.managed[c.addr.Bare().String()]
	return ok
}

// Leave exits the MUC, causing Joined to begin to return false.
func (c *Channel) Leave(ctx context.Context, status string) error {
	return c.LeavePresence(ctx, status, stanza.Presence{})
}

// LeavePresence is like Leave except that it gives you more control over the
// presence.
// Changing the presence type or to attributes have no effect.
func (c *Channel) LeavePresence(ctx context.Context, status string, p stanza.Presence) error {
	if p.Type != stanza.UnavailablePresence {
		p.Type = stanza.UnavailablePresence
	}
	if !p.To.Equal(c.addr) {
		p.To = c.addr
	}
	if p.ID == "" {
		p.ID = attr.RandomID()
	}

	var inner xml.TokenReader
	if status != "" {
		inner = xmlstream.Wrap(
			xmlstream.Token(xml.CharData(status)),
			xml.StartElement{Name: xml.Name{Local: "status"}},
		)
	}
	errChan := make(chan error)
	go func(errChan chan<- error) {
		resp, err := c.session.SendPresenceElement(ctx, inner, p)
		//err := s.Send(ctx, p.Wrap(conf.TokenReader()))
		if err != nil {
			errChan <- err
			return
		}
		/* #nosec */
		defer resp.Close()
		// Pop the start presence token.
		_, err = resp.Token()
		if err != nil {
			errChan <- err
			return
		}
		stanzaError, err := stanza.UnmarshalError(resp)
		if err != nil {
			errChan <- err
			return
		}
		errChan <- stanzaError
	}(errChan)

	select {
	case err := <-errChan:
		return err
	case <-c.depart:
	}
	return nil
}

// Invite sends a mediated invitation (an invitation sent from the channel
// itself) to the user.
//
// For direct invitations sent from your own account (ie. to avoid users who
// block all unrecognized JIDs) see the Invite function.
func (c *Channel) Invite(ctx context.Context, reason string, to jid.JID) error {
	return c.session.Send(ctx, stanza.Message{
		To:   c.addr.Bare(),
		Type: stanza.NormalMessage,
	}.Wrap(Invitation{
		JID:      to,
		Password: c.pass,
		Reason:   reason,
	}.MarshalMediated()))
}
