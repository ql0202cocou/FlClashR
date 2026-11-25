package interop

import (
	"bufio"
	"bytes"
	"context"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/mldsa"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	app "core"
	"github.com/metacubex/mihomo/adapter"
	"github.com/metacubex/mihomo/adapter/provider"
	C "github.com/metacubex/mihomo/constant"
)

func signedRealityTarget(t *testing.T) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	// REALITY needs a TLS certificate flight large enough to carry the ML-DSA signature.
	cert := &x509.Certificate{SerialNumber: big.NewInt(1), DNSNames: []string{"example.com"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		ExtraExtensions: []pkix.Extension{{Id: []int{1, 3, 6, 1, 4, 1, 32473, 1}, Value: make([]byte, 5200)}}}
	der, err := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	s := httptest.NewUnstartedServer(http.NotFoundHandler())
	s.EnableHTTP2 = true
	s.Config.ErrorLog = log.New(io.Discard, "", 0)
	s.TLS = &tls.Config{MinVersion: tls.VersionTLS13, CurvePreferences: []tls.CurveID{tls.X25519MLKEM768, tls.X25519}, Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}}}
	s.StartTLS()
	t.Cleanup(s.Close)
	return s.Listener.Addr().String()
}

type shareClient struct{ net.Conn }

func startShareClient(t *testing.T) shareClient {
	t.Helper()
	listener, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	listener.SetDeadline(time.Now().Add(5 * time.Second))
	done := make(chan struct{})
	go func() { defer close(done); app.StartServer(strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)) }()
	conn, err := listener.Accept()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close(); <-done })
	return shareClient{conn}
}

func (client shareClient) call(t *testing.T, method string, args, result any) {
	t.Helper()
	request, err := json.Marshal(map[string]any{"method": "shareLink." + method, "arguments": args})
	if err != nil {
		t.Fatal(err)
	}
	client.SetDeadline(time.Now().Add(5 * time.Second))
	if err := binary.Write(client, binary.LittleEndian, uint32(len(request))); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Write(request); err != nil {
		t.Fatal(err)
	}
	var size uint32
	if err := binary.Read(client, binary.LittleEndian, &size); err != nil {
		t.Fatal(err)
	}
	if size > 1<<20 {
		t.Fatalf("unexpected reply size %d", size)
	}
	data := make([]byte, size)
	if _, err := io.ReadFull(client, data); err != nil {
		t.Fatal(err)
	}
	var reply struct {
		Result json.RawMessage
		Error  *app.MethodError
	}
	if err := json.Unmarshal(data, &reply); err != nil {
		t.Fatal(err)
	}
	if reply.Error != nil {
		t.Fatal(reply.Error)
	}
	if err := json.Unmarshal(reply.Result, result); err != nil {
		t.Fatal(err)
	}
}

func (client shareClient) decodeNode(t *testing.T, link string) map[string]any {
	t.Helper()
	var groups [][]map[string]any
	client.call(t, "decode", app.LinesParams{Lines: []string{link}}, &groups)
	if len(groups) != 1 || len(groups[0]) != 1 {
		t.Fatal("expected one imported node")
	}
	return groups[0][0]
}

// A listening Xray socket can precede its first successful connection to the TLS target.
func waitForReality(t *testing.T, node map[string]any, echoURL string, port int) {
	t.Helper()
	proxy, err := adapter.ParseProxy(node)
	if err != nil {
		t.Fatal(err)
	}
	defer proxy.Close()
	for deadline := time.Now().Add(20 * time.Second); time.Now().Before(deadline); {
		err = probeEcho(proxy, echoURL, port, []byte("ready"))
		if err == nil {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("Xray REALITY readiness probe: %v", err)
}

func loadFileProvider(t *testing.T, link string) C.Proxy {
	t.Helper()
	const filename = "provider.txt"
	if err := os.WriteFile(filepath.Join(C.Path.HomeDir(), filename), []byte(link), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := provider.ParseProxyProvider("probe", map[string]any{"type": "file", "path": filename}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { p.(io.Closer).Close() })
	if err := p.Initial(); err != nil {
		t.Fatal(err)
	}
	if p.Count() != 1 {
		t.Fatalf("provider returned %d proxies", p.Count())
	}
	return p.Proxies()[0]
}

func probeEcho(proxy C.Proxy, echoURL string, port int, payload []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := proxy.DialContext(ctx, &C.Metadata{NetWork: C.TCP, DstIP: netip.MustParseAddr("127.0.0.1"), DstPort: uint16(port)})
	if err != nil {
		return err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	req, err := http.NewRequest("POST", echoURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Close = true
	if err = req.Write(conn); err != nil {
		return err
	}
	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	got, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK || !bytes.Equal(got, payload) {
		return fmt.Errorf("echo returned HTTP %d and %d bytes", resp.StatusCode, len(got))
	}
	return nil
}

// From app/core with Go 1.27+: go test ../../tests/interop/reality{,_pqv}_test.go -run TestRealityPQV -v -args -xray /path/to/xray
func TestRealityPQV(t *testing.T) {
	if *xray == "" {
		t.Skip("supply the official Xray v26.11.9 executable with -xray")
	}
	version, err := exec.Command(*xray, "version").Output()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(version), "Xray 26.11.9 ") {
		t.Fatalf("expected Xray 26.11.9, got %s", version)
	}
	client := startShareClient(t)
	target := signedRealityTarget(t)
	enc := base64.RawURLEncoding.EncodeToString
	xKey, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := mldsa.GenerateKey(mldsa.MLDSA65())
	if err != nil {
		t.Fatal(err)
	}
	other, err := mldsa.GenerateKey(mldsa.MLDSA65())
	if err != nil {
		t.Fatal(err)
	}
	var count atomic.Int32
	echo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count.Add(1)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Write(body)
	}))
	defer echo.Close()
	port := echo.Listener.Addr().(*net.TCPAddr).Port
	oldHome := C.Path.HomeDir()
	C.SetHomeDir(t.TempDir())
	defer C.SetHomeDir(oldHome)
	for _, profile := range []string{"tcp", "xhttp", "tcp-unsigned"} {
		network, seed := profile, enc(signer.Bytes())
		if profile == "tcp-unsigned" {
			network, seed = "tcp", ""
		}
		serverPort := freePort(t)
		flow := ""
		if network == "tcp" {
			flow = "xtls-rprx-vision"
		}
		serverConfig := fmt.Sprintf(`{
			"inbounds":[{"listen":"127.0.0.1","port":%d,"protocol":"vless",
			"settings":{"clients":[{"id":%q,"flow":%q}],"decryption":"none"},
			"streamSettings":{"network":%q,"security":"reality","xhttpSettings":{"path":"/interop"},
			"realitySettings":{"target":%q,"serverNames":["example.com"],"privateKey":%q,"shortIds":["01"],"mldsa65Seed":%q}}}],
			"outbounds":[{"protocol":"freedom","settings":{"redirect":"127.0.0.1:%d","finalRules":[{"action":"allow","network":"tcp","ip":["127.0.0.1/32"],"port":%d}]}}]
		}`, serverPort, userID, flow, network, target, enc(xKey.Bytes()), seed, port, port)
		startProxy(t, *xray, serverConfig, serverPort, "run", "-c", "config.json")
		q := url.Values{"security": {"reality"}, "fp": {"chrome"}, "pbk": {enc(xKey.PublicKey().Bytes())}, "sid": {"01"}, "sni": {"example.com"}, "type": {network}, "encryption": {"none"}}
		if network == "tcp" {
			q.Set("flow", flow)
		} else {
			q.Set("path", "/interop")
		}
		link := url.URL{Scheme: "vless", User: url.User(userID), Host: net.JoinHostPort("127.0.0.1", strconv.Itoa(serverPort)), Fragment: "probe", RawQuery: q.Encode()}
		waitForReality(t, client.decodeNode(t, link.String()), echo.URL, port)
		variants := []string{profile}
		if network == "xhttp" {
			variants = append(variants, "xhttp-download")
		}
		for _, variant := range variants {
			for _, tc := range []struct{ name, key string }{
				{"correct", enc(signer.PublicKey().Bytes())}, {"wrong", enc(other.PublicKey().Bytes())}, {"omitted", ""},
			} {
				t.Run(variant+"/"+tc.name, func(t *testing.T) {
					pub := enc(xKey.PublicKey().Bytes())
					q := link.Query()
					if variant == "xhttp-download" {
						q.Set("mode", "stream-up")
						q.Set("pqv", enc(signer.PublicKey().Bytes()))
						reality := map[string]any{"password": pub, "shortId": "01", "serverName": "example.com", "fingerprint": "chrome"}
						if tc.key != "" {
							reality["mldsa65Verify"] = tc.key
						}
						extra, err := json.Marshal(map[string]any{"downloadSettings": map[string]any{
							"address": "127.0.0.1", "port": serverPort, "network": "xhttp",
							"security": "reality", "realitySettings": reality,
							"xhttpSettings": map[string]any{"path": "/interop"},
						}})
						if err != nil {
							t.Fatal(err)
						}
						q.Set("extra", string(extra))
					} else if tc.key != "" {
						q.Set("pqv", tc.key)
					}
					u := link
					u.RawQuery = q.Encode()
					var exports []string
					client.call(t, "encode", app.ProxiesParams{Proxies: []map[string]any{client.decodeNode(t, u.String())}}, &exports)
					if len(exports) != 1 || exports[0] == "" {
						t.Fatal("shareLink.encode returned no link")
					}
					node := client.decodeNode(t, exports[0])
					r := node["reality-opts"].(map[string]any)
					if variant == "xhttp-download" {
						if r["mldsa65-verify"] != q.Get("pqv") {
							t.Fatal("uplink key changed on round-trip")
						}
						ds := node["xhttp-opts"].(map[string]any)["download-settings"].(map[string]any)
						r = ds["reality-opts"].(map[string]any)
					}
					value, present := r["mldsa65-verify"]
					if present != (tc.key != "") || present && value != tc.key {
						t.Fatal("verification key changed on round-trip")
					}
					for _, source := range []string{"inline", "provider"} {
						t.Run(source, func(t *testing.T) {
							var proxy C.Proxy
							if source == "provider" {
								proxy = loadFileProvider(t, exports[0])
							} else {
								inline, err := adapter.ParseProxy(node)
								if err != nil {
									t.Fatal(err)
								}
								proxy = inline
							}
							defer proxy.Close()
							before := count.Load()
							err := probeEcho(proxy, echo.URL, port, bytes.Repeat([]byte("FlClashR-loopback."), 100))
							hits := count.Load() - before
							if tc.name == "wrong" || (seed == "" && tc.key != "") {
								if err == nil || hits != 0 {
									t.Fatalf("expected signing failure: %v, %d echo requests", err, hits)
								}
								if network == "tcp" && !strings.Contains(err.Error(), "ML-DSA-65") {
									t.Fatalf("unexpected rejection: %v", err)
								}
								return
							}
							if err != nil || hits != 1 {
								t.Fatalf("echo: %v, %d requests", err, hits)
							}
						})
					}
				})
			}
		}
	}
}
