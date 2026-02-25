// internal/executor/executor.go
package executor

import (
	"context"
	"os"
)

type Executor interface {
	Run(ctx context.Context, cmd string) (string, error)
	WriteFile(ctx context.Context, path string, content []byte, mode os.FileMode) error
	ReadFile(ctx context.Context, path string) ([]byte, error)
}
