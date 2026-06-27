// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package helper

import (
	"os"
	"path/filepath"
)

// DefaultSocketPath is where the helper listens and the daemon looks. The
// HEIMDALL_HELPER_SOCKET environment variable overrides it; otherwise it lives
// under the system temp dir so unprivileged local development works without a
// root-owned /var/run path. Production deployments should pass an explicit
// --socket under a root-owned directory and restrict access by group.
func DefaultSocketPath() string {
	if p := os.Getenv("HEIMDALL_HELPER_SOCKET"); p != "" {
		return p
	}
	return filepath.Join(os.TempDir(), "heimdall-helper.sock")
}
