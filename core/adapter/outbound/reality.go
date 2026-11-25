package outbound

import (
	"crypto/ecdh"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	tlsC "github.com/metacubex/mihomo/component/tls"

	"github.com/samber/lo"
)

type RealityOptions struct {
	PublicKey     string `proxy:"public-key"`
	ShortID       string `proxy:"short-id,omitempty"`
	MLDSA65Verify string `proxy:"mldsa65-verify,omitempty"`

	SupportX25519MLKEM768 *bool `proxy:"support-x25519mlkem768,omitempty"`
}

func (o RealityOptions) Parse() (*tlsC.RealityConfig, error) {
	if o.MLDSA65Verify != "" && o.PublicKey == "" {
		return nil, errors.New("REALITY mldsa65-verify requires public-key")
	}
	if o.PublicKey != "" {
		config := new(tlsC.RealityConfig)
		config.SupportX25519MLKEM768 = lo.FromPtrOr(o.SupportX25519MLKEM768, true)
		if o.MLDSA65Verify != "" {
			verify, err := tlsC.NewRealityMLDSA65Verifier(o.MLDSA65Verify)
			if err != nil {
				return nil, err
			}
			config.MLDSA65Verify = verify
		}

		const x25519ScalarSize = 32
		publicKey, err := base64.RawURLEncoding.DecodeString(o.PublicKey)
		if err != nil || len(publicKey) != x25519ScalarSize {
			return nil, errors.New("invalid REALITY public key")
		}
		config.PublicKey, err = ecdh.X25519().NewPublicKey(publicKey)
		if err != nil {
			return nil, fmt.Errorf("fail to create REALITY public key: %w", err)
		}

		n := hex.DecodedLen(len(o.ShortID))
		if n > tlsC.RealityMaxShortIDLen {
			return nil, errors.New("invalid REALITY short id")
		}
		n, err = hex.Decode(config.ShortID[:], []byte(o.ShortID))
		if err != nil || n > tlsC.RealityMaxShortIDLen {
			return nil, errors.New("invalid REALITY short ID")
		}

		return config, nil
	}
	return nil, nil
}

func (o RealityOptions) parseForNetwork(network string, alpn []string) (*tlsC.RealityConfig, error) {
	if o.MLDSA65Verify != "" {
		switch {
		case network == "ws":
			return nil, errors.New("REALITY mldsa65-verify is not supported on the WebSocket transport")
		case network == "xhttp" && len(alpn) == 1 && alpn[0] == "h3":
			return nil, errors.New("XHTTP over HTTP/3 does not support REALITY mldsa65-verify")
		}
	}
	return o.Parse()
}
