// Package secrets env_provider: reads secrets from environment variables.
//
// This is the simplest provider and is used in development. No caching —
// every call reads from os.Getenv, so secrets can be rotated by changing
// the environment and restarting the process.
package secrets

import (
	"fmt"
	"os"
)

// EnvProvider reads secrets from environment variables.
type EnvProvider struct{}

// NewEnvProvider creates a new EnvProvider.
func NewEnvProvider() *EnvProvider {
	return &EnvProvider{}
}

// Get returns the value of the environment variable with the given key.
func (EnvProvider) Get(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("secret %q not found in environment", key)
	}
	return v, nil
}

// MustGet returns the value of the environment variable with the given key.
// Panics if not found.
func (p *EnvProvider) MustGet(key string) string {
	v, err := p.Get(key)
	if err != nil {
		panic(err)
	}
	return v
}
