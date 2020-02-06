// Copyright 2020 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package ibb

import (
	"errors"
	"net"

	"mellium.im/xmpp/jid"
)

var (
	ErrListenerClosed = errors.New("ibb: listener was closed")
)

type listener struct {
	conn chan net.Conn
	addr jid.JID
}

func (l *listener) Accept() (net.Conn, error) {
	conn, ok := <-l.conn
	if !ok {
		return nil, ErrListenerClosed
	}
	return conn, nil
}

func (l *listener) Close() error {
	close(l.conn)
	return nil
}

func (l *listener) Addr() net.Addr {
	return l.addr
}
