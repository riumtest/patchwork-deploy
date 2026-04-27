package ssh

import (
	"fmt"
	"io"
	"time"

	"golang.org/x/crypto/ssh"
)

// Config holds SSH connection parameters.
type Config struct {
	Address    string
	User       string
	PrivateKey []byte
	Timeout    time.Duration
}

// Executor runs scripts on a remote host over SSH.
type Executor struct {
	client *ssh.Client
}

// NewExecutor establishes an SSH connection and returns an Executor.
func NewExecutor(cfg Config) (*Executor, error) {
	signer, err := ssh.ParsePrivateKey(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	clientCfg := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // nolint: gosec
		Timeout:         timeout,
	}

	client, err := ssh.Dial("tcp", cfg.Address, clientCfg)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", cfg.Address, err)
	}

	return &Executor{client: client}, nil
}

// RunScript uploads and executes a shell script on the remote host.
// It returns combined stdout+stderr output and any error.
func (e *Executor) RunScript(script string) (string, error) {
	session, err := e.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("new session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(script)
	if err != nil {
		return string(output), fmt.Errorf("run script: %w", err)
	}

	return string(output), nil
}

// Close terminates the underlying SSH connection.
func (e *Executor) Close() error {
	return e.client.Close()
}

// RunReader executes the content provided by the reader as a shell script.
func (e *Executor) RunReader(r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("read script: %w", err)
	}
	return e.RunScript(string(data))
}
