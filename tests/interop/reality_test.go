package interop

import (
	"bytes"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

var (
	mihomo = flag.String("mihomo", "", "path to the mihomo executable under test")
	xray   = flag.String("xray", "", "path to an official Xray 24.12.31, 25.9.11 or 26.11.9 executable")
)

const userID = "11a962d3-82f0-47c5-bf65-18b0f40d4973"

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func startProxy(t *testing.T, binary, config string, port int, args ...string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(dir, "process.log")
	log, err := os.Create(logPath)
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	path, err := filepath.Abs(binary)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(path, args...)
	cmd.Dir = dir
	cmd.Stdout, cmd.Stderr = log, log
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		cmd.Wait()
		close(done)
	}()
	t.Cleanup(func() {
		cmd.Process.Kill()
		<-done
		if t.Failed() {
			output, _ := os.ReadFile(logPath)
			t.Logf("%s: %s", filepath.Base(binary), output)
		}
	})
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); {
		select {
		case <-done:
			t.Fatalf("%s exited before listening", binary)
		default:
		}
		if c, err := net.DialTimeout("tcp4", addr, 100*time.Millisecond); err == nil {
			c.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("%s did not listen on %s", binary, addr)
}

// The CLI starts its listeners before it finishes loading its routing table.
func waitForControlTunnel(t *testing.T, client *http.Client, port int) {
	t.Helper()
	url := fmt.Sprintf("http://127.0.0.1:%d/ready", port)
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusNoContent {
				return
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("control tunnel never reached the echo endpoint")
}

// Run from the repository root with Go 1.24 or newer:
// go test tests/interop/reality_test.go -v -args -mihomo /path/to/mihomo -xray /path/to/xray
func TestRealityMLKEM(t *testing.T) {
	if *xray == "" && *mihomo == "" {
		t.Skip("supply -mihomo and -xray to run the loopback test")
	}
	if *xray == "" || *mihomo == "" {
		t.Fatal("both -mihomo and -xray are required")
	}
	output, err := exec.Command(*xray, "version").Output()
	if err != nil {
		t.Fatal(err)
	}
	fields := strings.Fields(string(output))
	if len(fields) < 2 || fields[0] != "Xray" {
		t.Fatalf("unexpected Xray version output: %q", output)
	}
	version := fields[1]
	switch version {
	case "24.12.31", "25.9.11", "26.11.9":
	default:
		t.Fatalf("no expected handshake results for Xray %s", version)
	}
	key, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	var echoCount atomic.Int32
	echo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ready" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		echoCount.Add(1)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Write(body)
	}))
	defer echo.Close()
	echoPort := echo.Listener.Addr().(*net.TCPAddr).Port

	profiles := []string{"hybrid"}
	if version == "24.12.31" {
		profiles = append(profiles, "classic")
	}
	for _, profile := range profiles {
		t.Run(version+"/"+profile, func(t *testing.T) {
			curves := []tls.CurveID{tls.X25519MLKEM768, tls.X25519}
			if profile == "classic" {
				curves = []tls.CurveID{tls.X25519}
			}
			target := httptest.NewUnstartedServer(http.NotFoundHandler())
			target.EnableHTTP2 = true
			target.Config.ErrorLog = log.New(io.Discard, "", 0)
			target.TLS = &tls.Config{MinVersion: tls.VersionTLS13, CurvePreferences: curves}
			target.StartTLS()
			defer target.Close()
			serverPort := freePort(t)
			serverConfig := fmt.Sprintf(`{
				"inbounds":[{"listen":"127.0.0.1","port":%d,"protocol":"vless",
					"settings":{"clients":[{"id":%q,"flow":"xtls-rprx-vision"}],"decryption":"none"},
					"streamSettings":{"network":"tcp","security":"reality","realitySettings":{
						"dest":%q,"serverNames":["example.com"],"privateKey":%q,"shortIds":["01"]}}}],
				"outbounds":[{"protocol":"freedom","settings":{"redirect":"127.0.0.1:%d","finalRules":[
					{"action":"allow","network":"tcp","ip":["127.0.0.1/32"],"port":%d}]}}]
			}`, serverPort, userID, target.Listener.Addr().String(), base64.RawURLEncoding.EncodeToString(key.Bytes()), echoPort, echoPort)
			startProxy(t, *xray, serverConfig, serverPort, "run", "-c", "config.json")

			for _, option := range []string{"omitted", "true", "false"} {
				t.Run(option, func(t *testing.T) {
					setting := ""
					if option != "omitted" {
						setting = `, "support-x25519mlkem768":` + option
					}
					clientPort := freePort(t)
					readyPort := freePort(t)
					clientConfig := fmt.Sprintf(`{
						"tunnels":["tcp,127.0.0.1:%d,127.0.0.1:%d,probe","tcp,127.0.0.1:%d,127.0.0.1:%d,DIRECT"],
						"proxies":[{"name":"probe","type":"vless","server":"127.0.0.1","port":%d,
							"uuid":%q,"flow":"xtls-rprx-vision","tls":true,"servername":"example.com",
							"client-fingerprint":"chrome","reality-opts":{"public-key":%q,"short-id":"01"%s}}]
					}`, clientPort, echoPort, readyPort, echoPort, serverPort, userID, base64.RawURLEncoding.EncodeToString(key.PublicKey().Bytes()), setting)
					startProxy(t, *mihomo, clientConfig, readyPort, "-d", ".", "-f", "config.json")

					client := &http.Client{Timeout: 5 * time.Second}
					defer client.CloseIdleConnections()
					waitForControlTunnel(t, client, readyPort)
					payload := bytes.Repeat([]byte("FlClashR-loopback."), 100)
					before := echoCount.Load()
					resp, err := client.Post(fmt.Sprintf("http://127.0.0.1:%d/", clientPort), "application/octet-stream", bytes.NewReader(payload))
					want := true
					switch version {
					case "24.12.31":
						want = profile == "classic" || option == "false"
					case "26.11.9":
						want = option != "false"
					}
					if resp != nil {
						defer resp.Body.Close()
					}
					if !want {
						if echoCount.Load() != before || (err == nil && resp.StatusCode < 400) {
							t.Fatal("expected handshake rejection before the echo endpoint")
						}
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					got, err := io.ReadAll(resp.Body)
					if err != nil || resp.StatusCode != 200 || !bytes.Equal(got, payload) || echoCount.Load() != before+1 {
						t.Fatalf("echo: HTTP %d, received %d of %d bytes, error %v", resp.StatusCode, len(got), len(payload), err)
					}
				})
			}
		})
	}
}
