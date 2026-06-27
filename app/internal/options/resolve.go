// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package options

import (
	"flag"
	"fmt"
	"os"
)

// Resolve is the convenience entry point a binary calls after registering its
// catalog on flag.CommandLine and parsing: it loads the saved config, folds
// defaults < file < env < flags, runs the first-run wizard when nothing was
// provided on a terminal, persists when the wizard ran or a setting flag was
// given, and returns the effective settings. intro is the wizard banner — the
// first line is the title, the rest are notes.
func Resolve(binary string, cat Catalog, intro ...string) Resolved {
	path, _ := Locate(binary)
	fileSrc, found, _ := FromFile(path)

	resolver := NewResolver(cat).
		With(Builtins(cat)).
		With(fileSrc).
		With(FromEnvironment(cat)).
		With(FromFlags(cat, flag.CommandLine))
	resolved := resolver.Resolve()

	wizardRan := false
	if !found && !AnyProvided(flag.CommandLine) && Interactive() {
		prompter := NewPrompter(os.Stdin, os.Stderr)
		if len(intro) > 0 {
			prompter.Intro(intro[0], intro[1:]...)
		}
		resolved = resolver.With(RunWizard(cat, resolved, prompter)).Resolve()
		wizardRan = true
	}

	if path != "" && (wizardRan || Provided(cat, flag.CommandLine)) {
		if err := (Sink{Path: path}).Write(cat, resolved); err == nil {
			fmt.Fprintf(os.Stderr, "heimdall-%s: saved config to %s\n", binary, path)
		}
	}
	return resolved
}
