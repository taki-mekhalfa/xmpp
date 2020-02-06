// Copyright 2020 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package ibb_test

import (
	"context"
	"io"
	"strconv"
	"strings"
	"testing"

	"mellium.im/xmpp/ibb"
	"mellium.im/xmpp/internal/xmpptest"
	"mellium.im/xmpp/mux"
)

var (
	_ mux.IQHandler      = (*ibb.Handler)(nil)
	_ mux.MessageHandler = (*ibb.Handler)(nil)
)

const sendData = "To sit in solemn silence on a dull dark dock"

var sendDataTestCases = [...]struct {
	BlockSize uint16
}{
	0: {},
}

func TestSendData(t *testing.T) {
	for i, tc := range sendDataTestCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			serverHandler := &ibb.Handler{}
			serverMux := mux.New(ibb.Handle(serverHandler))
			clientHandler := &ibb.Handler{}
			clientMux := mux.New(ibb.Handle(clientHandler))
			cs := xmpptest.NewClientServer(
				xmpptest.ServerHandler(serverMux),
				xmpptest.ClientHandler(clientMux),
			)

			errChan := make(chan error, 2)
			go func() {
				conn, err := clientHandler.Open(context.Background(), cs.Client, cs.Server.LocalAddr(), tc.BlockSize)
				if err != nil {
					errChan <- err
					return
				}
				_, err = io.WriteString(conn, sendData)
				if err != nil {
					errChan <- err
				}
			}()
			var buf strings.Builder
			go func() {
				listener := serverHandler.Listen(cs.Server.LocalAddr())
				conn, err := listener.Accept()
				if err != nil {
					errChan <- err
					return
				}
				_, err = io.Copy(&buf, conn)
				if err != nil {
					errChan <- err
				}
			}()

			for err := range errChan {
				if err != nil {
					t.Errorf("unexpected error on channel: %v", err)
				}
			}

			if s := buf.String(); s != sendData {
				t.Errorf("transmitted data was not correct: want=%q, got=%q", sendData, s)
			}
		})
	}
}
