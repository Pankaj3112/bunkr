// internal/executor/mock.go
package executor

import (
	"context"
	"fmt"
	"os"
)

type MockCall struct {
	Method string
	Args   []interface{}
}

type MockExecutor struct {
	Calls       []MockCall
	RunOutputs  map[string]string
	RunErrors   map[string]error
	Files       map[string][]byte
	ReadErrors  map[string]error
	WriteErrors map[string]error
}

func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		RunOutputs:  make(map[string]string),
		RunErrors:   make(map[string]error),
		Files:       make(map[string][]byte),
		ReadErrors:  make(map[string]error),
		WriteErrors: make(map[string]error),
	}
}

func (m *MockExecutor) Run(_ context.Context, cmd string) (string, error) {
	m.Calls = append(m.Calls, MockCall{Method: "Run", Args: []interface{}{cmd}})
	if err, ok := m.RunErrors[cmd]; ok {
		return "", err
	}
	if out, ok := m.RunOutputs[cmd]; ok {
		return out, nil
	}
	return "", nil
}

func (m *MockExecutor) WriteFile(_ context.Context, path string, content []byte, mode os.FileMode) error {
	m.Calls = append(m.Calls, MockCall{Method: "WriteFile", Args: []interface{}{path, content, mode}})
	if err, ok := m.WriteErrors[path]; ok {
		return err
	}
	m.Files[path] = content
	return nil
}

func (m *MockExecutor) ReadFile(_ context.Context, path string) ([]byte, error) {
	m.Calls = append(m.Calls, MockCall{Method: "ReadFile", Args: []interface{}{path}})
	if err, ok := m.ReadErrors[path]; ok {
		return nil, err
	}
	if data, ok := m.Files[path]; ok {
		return data, nil
	}
	return nil, fmt.Errorf("file not found: %s", path)
}
