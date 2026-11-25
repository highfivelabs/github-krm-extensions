// Copyright 2026 High Five Labs
// SPDX-License-Identifier: Apache-2.0

package extensions

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/nacl/box"

	"github.com/highfivelabs/krm-rest-runtime/descriptor"
	"github.com/highfivelabs/krm-rest-runtime/reconciler"
	"github.com/highfivelabs/krm-rest-runtime/rest"
)

// Hooks: every secret kind (those whose body has a write-only
// encrypted_value) encrypts spec.secretRefs.value with the public key of
// the secret's scope before it is sent.
var Hooks = func() map[string][]reconciler.Hook {
	f, err := descriptor.Parse(Types)
	if err != nil {
		panic(err)
	}
	out := map[string][]reconciler.Hook{}
	for _, k := range f.Kinds {
		if k.IsWriteOnly("encrypted_value") {
			out[k.Kind] = []reconciler.Hook{EncryptSecret}
		}
	}
	return out
}()

// EncryptSecret replaces body["value"] with encrypted_value and key_id. The
// scope's public key is at the secret's path with "public-key" in place of
// the secret's name.
func EncryptSecret(ctx context.Context, c *rest.Client, k *descriptor.Kind, params map[string]string, body map[string]any) error {
	plain, ok := body["value"].(string)
	if !ok {
		return nil // encrypted_value given directly, or nothing to set
	}
	delete(body, "value")
	tmpl := k.Read.Path[:strings.LastIndex(k.Read.Path, "/")] + "/public-key"
	path, err := rest.Expand(tmpl, params)
	if err != nil {
		return err
	}
	resp, err := c.Do(ctx, "GET", path, nil)
	if err != nil {
		return fmt.Errorf("public key: %w", err)
	}
	m, _ := resp.(map[string]any)
	keyID, _ := m["key_id"].(string)
	keyB64, _ := m["key"].(string)
	enc, err := Seal(plain, keyB64)
	if err != nil {
		return err
	}
	body["encrypted_value"] = enc
	body["key_id"] = keyID
	return nil
}

// Seal encrypts plain for a base64 Curve25519 public key (libsodium's
// crypto_box_seal), as GitHub's secrets API expects.
func Seal(plain, publicKeyB64 string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil || len(raw) != 32 {
		return "", fmt.Errorf("public key: not a base64 32-byte key")
	}
	var pk [32]byte
	copy(pk[:], raw)
	out, err := box.SealAnonymous(nil, []byte(plain), &pk, rand.Reader)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(out), nil
}
