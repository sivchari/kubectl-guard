// Package config provides configuration management for kubectl-guard.
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the guard configuration.
type Config struct {
	GuardedContexts []GuardedContext `yaml:"guardedContexts"`
}

// GuardedContext represents a protected Kubernetes context.
type GuardedContext struct {
	Name       string   `yaml:"name"`
	Namespaces []string `yaml:"namespaces,omitempty"` // empty means all namespaces
}

// DefaultPath returns the default config file path.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".kube", "guard.yaml"), nil
}

// Load loads the config from the default path.
func Load() (*Config, error) {
	path, err := DefaultPath()
	if err != nil {
		return nil, err
	}
	return LoadFrom(path)
}

// LoadFrom loads the config from the specified path.
func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save saves the config to the default path.
func (c *Config) Save() error {
	path, err := DefaultPath()
	if err != nil {
		return err
	}
	return c.SaveTo(path)
}

// SaveTo saves the config to the specified path.
func (c *Config) SaveTo(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// IsGuarded checks if the context is guarded.
func (c *Config) IsGuarded(context string) bool {
	for _, gc := range c.GuardedContexts {
		if gc.Name == context {
			return true
		}
	}
	return false
}

// IsNamespaceGuarded checks if the namespace in the context is guarded.
func (c *Config) IsNamespaceGuarded(context, namespace string) bool {
	for _, gc := range c.GuardedContexts {
		if gc.Name == context {
			// empty namespaces means all namespaces are guarded
			if len(gc.Namespaces) == 0 {
				return true
			}
			for _, ns := range gc.Namespaces {
				if ns == namespace {
					return true
				}
			}
		}
	}
	return false
}

// AddContext adds a context to the guarded list.
func (c *Config) AddContext(context string, namespaces []string) {
	// Check if already exists
	for i, gc := range c.GuardedContexts {
		if gc.Name == context {
			c.GuardedContexts[i].Namespaces = namespaces
			return
		}
	}
	c.GuardedContexts = append(c.GuardedContexts, GuardedContext{
		Name:       context,
		Namespaces: namespaces,
	})
}

// RemoveContext removes a context from the guarded list.
func (c *Config) RemoveContext(context string) bool {
	for i, gc := range c.GuardedContexts {
		if gc.Name == context {
			c.GuardedContexts = append(c.GuardedContexts[:i], c.GuardedContexts[i+1:]...)
			return true
		}
	}
	return false
}
