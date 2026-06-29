// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Command heimdall-cli is the machine- and AI-friendly fleet client: it queries a
// running hub over the same subscription a dashboard uses and prints stable JSON,
// so scripts, CI/CD, and agent harnesses can consume the fleet without the TUI.
package main

import (
	"fmt"
	"os"

	"heimdall/app/internal/cli"
	"heimdall/app/internal/selfupdate"
)

// version is the Heimdall build version, set via -ldflags "-X main.version=…".
var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "update":
			if err := selfupdate.Run("cli", version); err != nil {
				fmt.Fprintln(os.Stderr, "heimdall-cli:", err)
				os.Exit(1)
			}
			return
		case "--version", "version":
			fmt.Println("heimdall-cli", version)
			return
		}
	}
	cli.Run(os.Args[1:])
}
