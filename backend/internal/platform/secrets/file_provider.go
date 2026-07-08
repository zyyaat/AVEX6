// Package secrets file_provider: reads secrets from a JSON file.
//
// The JSON file is a simple map of key → value:
//
//	{
//	  "JWT_SECRET": "my-secret-value",
//	  "DATABASE_URL": "postgres://...",
//	  ...
//	}
//
// The file is read once at startup and cached in memory. To rotate secrets,
// update the file and restart the process.
//
// This provider is suitable for staging environments where secrets are
// mounted from a secrets manager (e.g. Kubernetes secrets, Docker secrets)
// as files. For production with hot-reload needs, use the Vault provider.
package secrets

import (
	"encoding/json"
	"fmt"
	"os"
)

// FileProvider reads secrets from a JSON file.
type FileProvider struct {
	data map[string]string
}

// NewFileProvider creates a new FileProvider by reading and parsing the JSON file.
func NewFileProvider(filePath string) (*FileProvider, error) {
	if filePath == "" {
		return nil, fmt.Errorf("secrets file path is empty")
	}

	raw, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read secrets file %q: %w", filePath, err)
	}

	var data map[string]string
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("parse secrets file %q: %w", filePath, err)
	}

	return &FileProvider{data: data}, nil
}

// Get returns the value of the secret with the given key.
func (p *FileProvider) Get(key string) (string, error) {
	v, ok := p.data[key]
	if !ok {
		return "", fmt.Errorf("secret %q not found in file", key)
	}
	return v, nil
}

// MustGet returns the value of the secret with the given key.
// Panics if not found.
func (p *FileProvider) MustGet(key string) string {
	v, err := p.Get(key)
	if err != nil {
		panic(err)
	}
	return v
}
