package outbound

import (
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
