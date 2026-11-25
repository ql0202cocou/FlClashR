//go:build !go1.27

package tls

import "errors"

func NewRealityMLDSA65Verifier(string) (RealitySignatureVerifier, error) {
	return nil, errors.New("REALITY mldsa65-verify requires a core built with Go 1.27 or newer")
}
