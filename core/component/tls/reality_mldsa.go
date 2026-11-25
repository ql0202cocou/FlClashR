//go:build go1.27

package tls

import (
	"crypto/mldsa"
	"encoding/base64"
	"fmt"
)

func NewRealityMLDSA65Verifier(encoded string) (RealitySignatureVerifier, error) {
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid REALITY mldsa65-verify: %w", err)
	}
	key, err := mldsa.NewPublicKey(mldsa.MLDSA65(), raw)
	if err != nil {
		return nil, fmt.Errorf("invalid REALITY mldsa65-verify: %w", err)
	}
	return func(message, signature []byte) error {
		return mldsa.Verify(key, message, signature, nil)
	}, nil
}
