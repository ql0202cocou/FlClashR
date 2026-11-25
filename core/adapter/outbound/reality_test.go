package outbound

import (
	"strings"
	"testing"

	"github.com/samber/lo"
)

func TestRealityMLKEMDefault(t *testing.T) {
	for _, tc := range []struct {
		name   string
		option *bool
		want   bool
	}{
		{"omitted", nil, true},
		{"enabled", lo.ToPtr(true), true},
		{"disabled", lo.ToPtr(false), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			options := RealityOptions{
				PublicKey:             "ppQ9FwLrLIa0AOrp1WvcyiaQ37vg2WSy_CD4bIdiTUw",
				SupportX25519MLKEM768: tc.option,
			}
			config, err := options.Parse()
			if err != nil {
				t.Fatal(err)
			}
			if config.SupportX25519MLKEM768 != tc.want {
				t.Errorf("SupportX25519MLKEM768 = %t, want %t", config.SupportX25519MLKEM768, tc.want)
			}
		})
	}
}

func TestRealityRejectsUnusableVerification(t *testing.T) {
	const pub = "ppQ9FwLrLIa0AOrp1WvcyiaQ37vg2WSy_CD4bIdiTUw"
	for _, tc := range []struct {
		name, publicKey, verify, network string
		alpn                             []string
		wantError                        string
	}{
		{"missing X25519 key", "", "value", "tcp", nil, "requires public-key"},
		{"malformed encoding", pub, "%", "tcp", nil, "mldsa65-verify"},
		{"short key", pub, "AA", "tcp", nil, "mldsa65-verify"},
		{"WebSocket", pub, "value", "ws", nil, "WebSocket"},
		{"HTTP/3", pub, "value", "xhttp", []string{"h3"}, "HTTP/3"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opts := RealityOptions{PublicKey: tc.publicKey, MLDSA65Verify: tc.verify}
			config, err := opts.parseForNetwork(tc.network, tc.alpn)
			if config != nil || err == nil || !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("expected %q, got %v", tc.wantError, err)
			}
		})
	}
}
