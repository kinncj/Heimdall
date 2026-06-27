// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package control is the read-only, allow-listed remote control plane. Logical
// command keys map to fixed argv vectors; user-supplied arguments are validated,
// never concatenated into a shell, and never escalated. Every invocation —
// allowed or refused — is audited.
package control

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// Command is one allow-listed, read-only operation. Argv is a fixed command and
// flags; user-supplied args are appended only after ArgValidator approves them.
type Command struct {
	Key         string
	Argv        []string
	Description string
	// ArgValidator validates request args and returns the validated argv tail.
	// Nil means the command accepts no arguments.
	ArgValidator func(args []string) ([]string, error)
}

// Allowlist maps logical keys to allowed commands. A request whose key is absent
// is refused; there is no wildcard and no shell.
type Allowlist map[string]Command

// Resolve returns the full argv to execute for key+args, or an error if the key
// is not allow-listed or the args fail validation. It never returns a shell
// string and never includes sudo.
func (a Allowlist) Resolve(key string, args []string) ([]string, error) {
	cmd, ok := a[key]
	if !ok {
		return nil, fmt.Errorf("command %q is not allow-listed", key)
	}
	argv := append([]string(nil), cmd.Argv...)
	if cmd.ArgValidator == nil {
		if len(args) > 0 {
			return nil, fmt.Errorf("command %q accepts no arguments", key)
		}
		return argv, nil
	}
	tail, err := cmd.ArgValidator(args)
	if err != nil {
		return nil, err
	}
	return append(argv, tail...), nil
}

// Keys lists the allow-listed command keys, sorted, for discovery.
func (a Allowlist) Keys() []string {
	ks := make([]string, 0, len(a))
	for k := range a {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

// DefaultAllowlist is the built-in set of read-only commands. All are safe,
// unprivileged, and produce bounded output.
func DefaultAllowlist() Allowlist {
	return Allowlist{
		"process.list": {Key: "process.list", Argv: []string{"ps", "-eo", "pid,ppid,pcpu,pmem,comm"}, Description: "list processes"},
		"disk.df":      {Key: "disk.df", Argv: []string{"df", "-h"}, Description: "disk usage"},
		"uptime":       {Key: "uptime", Argv: []string{"uptime"}, Description: "load average and uptime"},
		"dir.list": {
			Key: "dir.list", Argv: []string{"ls", "-la"}, Description: "list an allow-listed directory",
			ArgValidator: allowedDirArg,
		},
	}
}

// allowedDirRoots bounds dir.list to non-sensitive, operator-relevant trees.
var allowedDirRoots = []string{"/var/log", "/tmp"}

// allowedDirArg permits exactly one absolute path that resolves within an
// allow-listed root. It cleans the path first, so traversal that escapes the
// roots (../) is refused.
func allowedDirArg(args []string) ([]string, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("dir.list takes exactly one directory")
	}
	clean := filepath.Clean(args[0])
	if !filepath.IsAbs(clean) {
		return nil, fmt.Errorf("directory must be an absolute path")
	}
	for _, root := range allowedDirRoots {
		if clean == root || strings.HasPrefix(clean, root+string(filepath.Separator)) {
			return []string{clean}, nil
		}
	}
	return nil, fmt.Errorf("directory %q is not within an allow-listed root", clean)
}
