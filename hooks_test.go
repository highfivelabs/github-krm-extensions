// Copyright 2026 High Five Labs
// SPDX-License-Identifier: Apache-2.0

package extensions

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/crypto/nacl/box"

	"github.com/highfivelabs/krm-rest-runtime/descriptor"
	"github.com/highfivelabs/krm-rest-runtime/rest"
)

// The value sent is the plain value sealed to the scope's public key: the
// key's owner can open it, and the plain value never leaves the hook.
func TestEncryptSecret(t *testing.T) {
	pub, priv, _ := box.GenerateKey(rand.Reader)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orgs/acme/actions/secrets/public-key" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"key_id": "k1", "key": base64.StdEncoding.EncodeToString(pub[:])})
	}))
	defer srv.Close()
	f, _ := descriptor.Parse(Types)
	var k *descriptor.Kind
	for i := range f.Kinds {
		if f.Kinds[i].Kind == "ActionsOrganizationSecret" {
			k = &f.Kinds[i]
		}
	}
	if k == nil {
		t.Fatal("no ActionsOrganizationSecret kind")
	}
	if len(Hooks[k.Kind]) != 1 {
		t.Fatalf("ActionsOrganizationSecret has %d hooks, want 1", len(Hooks[k.Kind]))
	}
	body := map[string]any{"value": "hunter2", "visibility": "all"}
	c := rest.New(Vendor, srv.URL, map[string]string{"token": "t"})
	if err := EncryptSecret(context.Background(), c, k, map[string]string{"org": "acme", "secret_name": "S"}, body); err != nil {
		t.Fatal(err)
	}
	if _, ok := body["value"]; ok || body["key_id"] != "k1" {
		t.Fatalf("body = %v", body)
	}
	sealed, _ := base64.StdEncoding.DecodeString(body["encrypted_value"].(string))
	plain, ok := box.OpenAnonymous(nil, sealed, pub, priv)
	if !ok || string(plain) != "hunter2" {
		t.Fatalf("opened %q, %v", plain, ok)
	}
}
