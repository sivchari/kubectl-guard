// Package guard provides context protection functionality for kubectl-guard.
package guard

import (
	"os"
	"os/exec"
	"strings"

	"github.com/sivchari/kubectl-guard/internal/config"
)

// DestructiveCommands lists kubectl commands that modify or delete resources.
var DestructiveCommands = []string{
	"delete",
	"apply",
	"patch",
	"replace",
	"scale",
	"rollout",
	"drain",
	"cordon",
	"uncordon",
	"taint",
	"label",
	"annotate",
	"edit",
	"set",
}

// Guard provides context protection functionality.
type Guard struct {
	cfg *config.Config
}

// New creates a new Guard instance.
func New(cfg *config.Config) *Guard {
	return &Guard{cfg: cfg}
}

// CheckResult represents the result of a guard check.
type CheckResult struct {
	Blocked   bool
	Context   string
	Namespace string
	Command   string
	Message   string
}

// Check checks if the command should be blocked.
func (g *Guard) Check(args []string) (*CheckResult, error) {
	ctx, err := GetCurrentContext()
	if err != nil {
		return nil, err
	}

	ns := GetNamespaceFromArgs(args)
	if ns == "" {
		ns, _ = GetCurrentNamespace()
	}

	cmd := GetCommand(args)

	result := &CheckResult{
		Context:   ctx,
		Namespace: ns,
		Command:   cmd,
	}

	if !g.cfg.IsGuarded(ctx) {
		return result, nil
	}

	if !g.cfg.IsNamespaceGuarded(ctx, ns) {
		return result, nil
	}

	if !IsDestructiveCommand(cmd) {
		return result, nil
	}

	result.Blocked = true
	result.Message = formatBlockMessage(ctx, ns, cmd)
	return result, nil
}

// GetCurrentContext returns the current kubectl context.
func GetCurrentContext() (string, error) {
	cmd := exec.Command("kubectl", "config", "current-context")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// GetCurrentNamespace returns the current namespace from kubeconfig.
func GetCurrentNamespace() (string, error) {
	cmd := exec.Command("kubectl", "config", "view", "--minify", "-o", "jsonpath={..namespace}")
	out, err := cmd.Output()
	if err != nil {
		return "default", nil
	}
	ns := strings.TrimSpace(string(out))
	if ns == "" {
		return "default", nil
	}
	return ns, nil
}

// GetNamespaceFromArgs extracts namespace from kubectl args.
func GetNamespaceFromArgs(args []string) string {
	for i, arg := range args {
		if arg == "-n" || arg == "--namespace" {
			if i+1 < len(args) {
				return args[i+1]
			}
		}
		if strings.HasPrefix(arg, "-n=") {
			return strings.TrimPrefix(arg, "-n=")
		}
		if strings.HasPrefix(arg, "--namespace=") {
			return strings.TrimPrefix(arg, "--namespace=")
		}
	}
	return ""
}

// flagsWithValue lists flags that take a value argument.
var flagsWithValue = []string{
	"-n", "--namespace",
	"-l", "--selector",
	"-f", "--filename",
	"-o", "--output",
	"-c", "--container",
	"--context",
	"--kubeconfig",
	"--cluster",
	"--user",
}

// GetCommand extracts the main kubectl command from args.
func GetCommand(args []string) string {
	skipNext := false
	for _, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if strings.HasPrefix(arg, "-") {
			// Check if this flag takes a value
			for _, f := range flagsWithValue {
				if arg == f {
					skipNext = true
					break
				}
			}
			continue
		}
		return arg
	}
	return ""
}

// IsDestructiveCommand checks if the command is destructive.
func IsDestructiveCommand(cmd string) bool {
	for _, dc := range DestructiveCommands {
		if cmd == dc {
			return true
		}
	}
	return false
}

// HasForceFlag checks if --force flag is present.
func HasForceFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--force" {
			return true
		}
	}
	return false
}

// RemoveForceFlag removes --force flag from args.
func RemoveForceFlag(args []string) []string {
	result := make([]string, 0, len(args))
	for _, arg := range args {
		if arg != "--force" {
			result = append(result, arg)
		}
	}
	return result
}

// ExecKubectl executes kubectl with the given args.
func ExecKubectl(args []string) error {
	kubectlPath, err := exec.LookPath("kubectl")
	if err != nil {
		return err
	}
	return execCommand(kubectlPath, args)
}

func execCommand(path string, args []string) error {
	cmd := exec.Command(path, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func formatBlockMessage(ctx, ns, cmd string) string {
	return "blocked\n" +
		"  context: " + ctx + "\n" +
		"  namespace: " + ns + "\n" +
		"  command: " + cmd + "\n\n" +
		"This context is guarded.\n" +
		"Use --force flag to execute, or run `kubectl guard unguard " + ctx + "` to remove protection."
}
