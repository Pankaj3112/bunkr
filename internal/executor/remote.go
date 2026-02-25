// internal/executor/remote.go
package executor

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type RemoteExecutor struct {
	client *ssh.Client
}

func NewRemoteExecutor(target string) (*RemoteExecutor, error) {
	user, host := parseTarget(target)

	// Ensure host has port
	if _, _, err := net.SplitHostPort(host); err != nil {
		host = net.JoinHostPort(host, "22")
	}

	authMethods, err := buildAuthMethods()
	if err != nil {
		return nil, fmt.Errorf("failed to build SSH auth: %w", err)
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", target, err)
	}

	return &RemoteExecutor{client: client}, nil
}

func (r *RemoteExecutor) Run(_ context.Context, cmd string) (string, error) {
	session, err := r.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(cmd); err != nil {
		return "", fmt.Errorf("%w: %s", err, stderr.String())
	}
	return stdout.String(), nil
}

func (r *RemoteExecutor) WriteFile(_ context.Context, path string, content []byte, mode os.FileMode) error {
	session, err := r.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Ensure parent directory exists, then write via cat
	cmd := fmt.Sprintf("mkdir -p %s && cat > %s && chmod %o %s",
		filepath.Dir(path), path, mode, path)

	session.Stdin = bytes.NewReader(content)
	var stderr bytes.Buffer
	session.Stderr = &stderr

	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("failed to write %s: %w: %s", path, err, stderr.String())
	}
	return nil
}

func (r *RemoteExecutor) ReadFile(_ context.Context, path string) ([]byte, error) {
	session, err := r.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(fmt.Sprintf("cat %s", path)); err != nil {
		return nil, fmt.Errorf("failed to read %s: %w: %s", path, err, stderr.String())
	}
	return stdout.Bytes(), nil
}

func parseTarget(target string) (user, host string) {
	if idx := strings.Index(target, "@"); idx != -1 {
		return target[:idx], target[idx+1:]
	}
	return "root", target
}

func buildAuthMethods() ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod

	// Try SSH agent first
	if sock := os.Getenv("SSH_AUTH_SOCK"); sock != "" {
		conn, err := net.Dial("unix", sock)
		if err == nil {
			agentClient := agent.NewClient(conn)
			methods = append(methods, ssh.PublicKeysCallback(agentClient.Signers))
		}
	}

	// Try default key files
	home, err := os.UserHomeDir()
	if err != nil {
		if len(methods) > 0 {
			return methods, nil
		}
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}

	keyFiles := []string{
		filepath.Join(home, ".ssh", "id_ed25519"),
		filepath.Join(home, ".ssh", "id_rsa"),
	}

	for _, keyFile := range keyFiles {
		key, err := os.ReadFile(keyFile)
		if err != nil {
			continue
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			continue
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}

	if len(methods) == 0 {
		return nil, fmt.Errorf("no SSH authentication methods available (tried agent and key files)")
	}

	return methods, nil
}
