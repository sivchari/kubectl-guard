// Package cli provides the command-line interface for kubectl-guard.
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/sivchari/kubectl-guard/internal/config"
	"github.com/sivchari/kubectl-guard/internal/guard"
)

const usage = `kubectl-guard - Kubernetes context protection plugin

Usage:
  kubectl guard <command> [options]

Commands:
  guard <context> [--namespace=<ns>]  Protect a context
  unguard <context>                   Remove protection from a context
  list                                List protected contexts and current status
  exec -- <kubectl args>              Execute kubectl with protection check

Examples:
  kubectl guard guard prod-cluster
  kubectl guard guard prod-cluster --namespace=production
  kubectl guard unguard prod-cluster
  kubectl guard list
  kubectl guard exec -- delete pod nginx

Options:
  --force    Force execution on protected context
  --help     Show help
`

// Run executes the CLI.
func Run(args []string) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		fmt.Print(usage)
		return 0
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return 1
	}

	switch args[0] {
	case "guard":
		return runGuard(cfg, args[1:])
	case "unguard":
		return runUnguard(cfg, args[1:])
	case "list":
		return runList(cfg)
	case "exec":
		return runExec(cfg, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[0])
		fmt.Print(usage)
		return 1
	}
}

func runGuard(cfg *config.Config, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "context name is required")
		return 1
	}

	context := args[0]
	var namespaces []string

	for _, arg := range args[1:] {
		if strings.HasPrefix(arg, "--namespace=") {
			ns := strings.TrimPrefix(arg, "--namespace=")
			namespaces = strings.Split(ns, ",")
		} else if strings.HasPrefix(arg, "-n=") {
			ns := strings.TrimPrefix(arg, "-n=")
			namespaces = strings.Split(ns, ",")
		}
	}

	cfg.AddContext(context, namespaces)
	if err := cfg.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to save config: %v\n", err)
		return 1
	}

	if len(namespaces) > 0 {
		fmt.Printf("guarded %s (namespaces: %s)\n", context, strings.Join(namespaces, ", "))
	} else {
		fmt.Printf("guarded %s (all namespaces)\n", context)
	}
	return 0
}

func runUnguard(cfg *config.Config, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "context name is required")
		return 1
	}

	context := args[0]
	if !cfg.RemoveContext(context) {
		fmt.Fprintf(os.Stderr, "%s is not guarded\n", context)
		return 1
	}

	if err := cfg.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to save config: %v\n", err)
		return 1
	}

	fmt.Printf("unguarded %s\n", context)
	return 0
}

func runList(cfg *config.Config) int {
	ctx, _ := guard.GetCurrentContext()

	if len(cfg.GuardedContexts) == 0 {
		fmt.Println("no guarded contexts")
	} else {
		fmt.Println("guarded contexts:")
		for _, gc := range cfg.GuardedContexts {
			marker := " "
			if gc.Name == ctx {
				marker = "*"
			}
			if len(gc.Namespaces) > 0 {
				fmt.Printf(" %s %s (namespaces: %s)\n", marker, gc.Name, strings.Join(gc.Namespaces, ", "))
			} else {
				fmt.Printf(" %s %s (all namespaces)\n", marker, gc.Name)
			}
		}
	}

	fmt.Println()
	if ctx != "" {
		if cfg.IsGuarded(ctx) {
			fmt.Printf("current: %s (guarded)\n", ctx)
		} else {
			fmt.Printf("current: %s (not guarded)\n", ctx)
		}
	}
	return 0
}

func runExec(cfg *config.Config, args []string) int {
	// Remove "--" separator if present
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "kubectl command is required")
		return 1
	}

	g := guard.New(cfg)

	// Check for --force flag
	forceMode := guard.HasForceFlag(args)
	if forceMode {
		args = guard.RemoveForceFlag(args)
	}

	result, err := g.Check(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "check failed: %v\n", err)
		return 1
	}

	if result.Blocked && !forceMode {
		fmt.Fprintln(os.Stderr, result.Message)
		return 1
	}

	if forceMode && result.Blocked {
		fmt.Fprintf(os.Stderr, "executing %s on %s with --force\n", result.Command, result.Context)
	}

	if err := guard.ExecKubectl(args); err != nil {
		return 1
	}
	return 0
}
