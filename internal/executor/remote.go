// internal/executor/remote.go
package executor

import (
	"context"
	"fmt"
	"os"
)

type RemoteExecutor struct{}

func NewRemoteExecutor(target string) (*RemoteExecutor, error) {
	return nil, fmt.Errorf("remote execution not yet implemented (target: %s)", target)
}

func (r *RemoteExecutor) Run(_ context.Context, cmd string) (string, error) {
	return "", fmt.Errorf("remote execution not implemented")
}

func (r *RemoteExecutor) WriteFile(_ context.Context, path string, content []byte, mode os.FileMode) error {
	return fmt.Errorf("remote execution not implemented")
}

func (r *RemoteExecutor) ReadFile(_ context.Context, path string) ([]byte, error) {
	return nil, fmt.Errorf("remote execution not implemented")
}
