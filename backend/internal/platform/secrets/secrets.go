// Package secrets defines the Provider interface for secret management.
//
// The Provider abstracts where secrets come from (env vars, files, Vault, AWS SM).
// This allows the platform to start with env vars in development and migrate
// to Vault in production without changing code that consumes secrets.
//
// Usage:
//
//	provider, err := secrets.NewProvider(cfg.Secrets)
//	if err != nil { ... }
//	jwtSecret, err := provider.Get("JWT_SECRET")
package secrets

import "fmt"

// Provider provides access to secrets by key name.
type Provider interface {
	// Get retrieves the value of a secret by key.
	// Returns an error if the secret is not found.
	Get(key string) (string, error)

	// MustGet retrieves the value of a secret by key.
	// Panics if the secret is not found. Use only during startup
	// for secrets that are absolutely required.
	MustGet(key string) string
}

// NewProvider creates a Provider based on the provider type.
func NewProvider(providerType string, filePath string) (Provider, error) {
	switch providerType {
	case "env", "":
		return NewEnvProvider(), nil
	case "file":
		p, err := NewFileProvider(filePath)
		if err != nil {
			return nil, fmt.Errorf("create file secrets provider: %w", err)
		}
		return p, nil
	case "vault":
		return nil, fmt.Errorf("vault provider not yet implemented")
	default:
		return nil, fmt.Errorf("unknown secrets provider: %q", providerType)
	}
}
