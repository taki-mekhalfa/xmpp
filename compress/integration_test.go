// Copyright 2020 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

//+build integration

package compress_test

import (
	"context"
	"crypto/tls"
	"testing"

	"mellium.im/sasl"
	"mellium.im/xmpp"
	"mellium.im/xmpp/compress"
	"mellium.im/xmpp/internal/integration"
	"mellium.im/xmpp/internal/integration/ejabberd"
	"mellium.im/xmpp/internal/integration/prosody"
)

func TestIntegrationCompression(t *testing.T) {
	prosodyRun := prosody.Test(context.TODO(), t,
		//integration.Log(),
		prosody.ListenC2S(),
		prosody.StreamCompression(),
	)
	prosodyRun(integrationCompress)

	ejabberdRun := ejabberd.Test(context.TODO(), t,
		integration.Log(),
		ejabberd.ListenC2S(),
		//ejabberd.StreamCompression(),
	)
	ejabberdRun(integrationCompress)
}

func integrationCompress(ctx context.Context, t *testing.T, cmd *integration.Cmd) {
	j, pass := cmd.User()
	session, err := cmd.DialClient(ctx, j, t,
		compress.New(),
		xmpp.StartTLS(&tls.Config{
			InsecureSkipVerify: true,
		}),
		xmpp.SASL("", pass, sasl.Plain),
		xmpp.BindResource(),
	)
	if err != nil {
		t.Fatalf("error connecting: %v", err)
	}
	_, ok := session.Feature(compress.NSFeatures)
	if !ok {
		t.Fatal("stream compression was not negotiated")
	}
}
