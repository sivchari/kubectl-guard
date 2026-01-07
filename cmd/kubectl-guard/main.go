// Package main provides the kubectl-guard CLI entry point.
package main

import (
	"fmt"
	"os"

	"github.com/sivchari/kubectl-guard/internal/cli"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println("kubectl-guard", version)
		return
	}
	os.Exit(cli.Run(os.Args[1:]))
}
