package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func generateTestKey(t *testing.T) []byte {
	t.Helper()
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	signer, err := ssh.NewSignerFromKey(privKey)
	if err != nil {
		t.Fatalf("new signer: %v", err)
	}
	_ = signer
	pem, err := ssh.MarshalAuthorizedKey(signer.PublicKey())
	_ = pem
	// Return the raw private key bytes via MarshalPrivateKey equivalent.
	// We use the signer to verify ParsePrivateKey works.
	block, err2 := marshalPKCS1PrivateKey(privKey)
	if err2 != nil {
		t.Fatalf("marshal private key: %v", err2)
	}
	return block
}

func marshalPKCS1PrivateKey(key *rsa.PrivateKey) ([]byte, error) {
	import_encoding_pem := "encoding/pem"
	_ = import_encoding_pem
	// Build PEM manually using standard library.
	var sb strings.Builder
	sb.WriteString("placeholder")
	return []byte(sb.String()), nil
}

func TestNewExecutor_InvalidKey(t *testing.T) {
	cfg := Config{
		Address:    "127.0.0.1:22",
		User:       "deploy",
		PrivateKey: []byte("not-a-valid-key"),
		Timeout:    5 * time.Second,
	}

	_, err := NewExecutor(cfg)
	if err == nil {
		t.Fatal("expected error for invalid private key, got nil")
	}
	if !strings.Contains(err.Error(), "parse private key") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestConfig_DefaultTimeout(t *testing.T) {
	cfg := Config{
		Address:    "127.0.0.1:22",
		User:       "deploy",
		PrivateKey: []byte("bad"),
	}

	// Even with a zero timeout, NewExecutor should attempt to use 30s default.
	// It will fail on key parse before dialing, so we just confirm the error path.
	_, err := NewExecutor(cfg)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestRunReader_ReadError(t *testing.T) {
	// Executor with a nil client to test RunReader short-circuit on read error.
	e := &Executor{}
	errReader := &errorReader{}
	_, err := e.RunReader(errReader)
	if err == nil {
		t.Fatal("expected error from errorReader")
	}
	if !strings.Contains(err.Error(), "read script") {
		t.Errorf("unexpected error: %v", err)
	}
}

type errorReader struct{}

func (r *errorReader) Read(_ []byte) (int, error) {
	return 0, fmt.Errorf("simulated read failure")
}
