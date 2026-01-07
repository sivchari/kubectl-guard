package guard

import (
	"testing"

	"github.com/sivchari/kubectl-guard/internal/config"
)

func TestGetNamespaceFromArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "short flag with space",
			args:     []string{"delete", "pod", "-n", "production", "nginx"},
			expected: "production",
		},
		{
			name:     "short flag with equals",
			args:     []string{"delete", "pod", "-n=production", "nginx"},
			expected: "production",
		},
		{
			name:     "long flag with space",
			args:     []string{"delete", "pod", "--namespace", "production", "nginx"},
			expected: "production",
		},
		{
			name:     "long flag with equals",
			args:     []string{"delete", "pod", "--namespace=production", "nginx"},
			expected: "production",
		},
		{
			name:     "no namespace flag",
			args:     []string{"delete", "pod", "nginx"},
			expected: "",
		},
		{
			name:     "flag at end without value",
			args:     []string{"delete", "pod", "-n"},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetNamespaceFromArgs(tt.args)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "simple command",
			args:     []string{"delete", "pod", "nginx"},
			expected: "delete",
		},
		{
			name:     "with flags first",
			args:     []string{"-n", "production", "delete", "pod", "nginx"},
			expected: "delete",
		},
		{
			name:     "only flags",
			args:     []string{"-n", "production"},
			expected: "",
		},
		{
			name:     "empty args",
			args:     []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCommand(tt.args)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestIsDestructiveCommand(t *testing.T) {
	destructive := []string{
		"delete", "apply", "patch", "replace", "scale",
		"rollout", "drain", "cordon", "uncordon", "taint",
		"label", "annotate", "edit", "set",
	}
	for _, cmd := range destructive {
		if !IsDestructiveCommand(cmd) {
			t.Errorf("expected %q to be destructive", cmd)
		}
	}

	safe := []string{"get", "describe", "logs", "exec", "port-forward", "top"}
	for _, cmd := range safe {
		if IsDestructiveCommand(cmd) {
			t.Errorf("expected %q to be safe", cmd)
		}
	}
}

func TestHasForceFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "with force flag",
			args:     []string{"delete", "--force", "pod", "nginx"},
			expected: true,
		},
		{
			name:     "without force flag",
			args:     []string{"delete", "pod", "nginx"},
			expected: false,
		},
		{
			name:     "force flag at end",
			args:     []string{"delete", "pod", "nginx", "--force"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasForceFlag(tt.args)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRemoveForceFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "with force flag",
			args:     []string{"delete", "--force", "pod", "nginx"},
			expected: []string{"delete", "pod", "nginx"},
		},
		{
			name:     "without force flag",
			args:     []string{"delete", "pod", "nginx"},
			expected: []string{"delete", "pod", "nginx"},
		},
		{
			name:     "multiple force flags",
			args:     []string{"--force", "delete", "--force", "pod"},
			expected: []string{"delete", "pod"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveForceFlag(tt.args)
			if len(result) != len(tt.expected) {
				t.Fatalf("expected length %d, got %d", len(tt.expected), len(result))
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("at index %d: expected %q, got %q", i, tt.expected[i], result[i])
				}
			}
		})
	}
}

func TestGuard_Check(t *testing.T) {
	cfg := &config.Config{
		GuardedContexts: []config.GuardedContext{
			{Name: "prod"},
			{Name: "staging", Namespaces: []string{"critical"}},
		},
	}
	g := New(cfg)

	// Note: These tests would require mocking kubectl commands
	// For now, we just verify the Guard struct is properly initialized
	if g.cfg != cfg {
		t.Error("Guard config not properly set")
	}
}
