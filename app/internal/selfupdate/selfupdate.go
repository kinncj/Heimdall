// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package selfupdate replaces the running binary with the latest published
// GitHub release asset for this OS/arch, after verifying its SHA-256 against the
// release's SHA256SUMS manifest. It depends only on the standard library.
package selfupdate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Repo is the owner/name queried for releases; override with $HEIMDALL_REPO.
func repo() string {
	if r := os.Getenv("HEIMDALL_REPO"); r != "" {
		return r
	}
	return "kinncj/Heimdall"
}

var client = &http.Client{Timeout: 60 * time.Second}

// assetName is the release asset for a binary on the current platform, matching
// scripts/release.sh, e.g. "heimdall-daemon_darwin_arm64".
func assetName(binary string) string {
	name := fmt.Sprintf("heimdall-%s_%s_%s", binary, runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return name
}

// sumFor returns the hex SHA-256 for asset from a SHA256SUMS manifest body.
func sumFor(manifest, asset string) (string, bool) {
	for _, line := range strings.Split(manifest, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == asset {
			return fields[0], true
		}
	}
	return "", false
}

// Run updates the named binary in place. current is its build version; when it
// already matches the latest release the update is skipped.
func Run(binary, current string) error {
	tag, err := latestTag()
	if err != nil {
		return err
	}
	if tag == current {
		fmt.Printf("heimdall-%s is already up to date (%s)\n", binary, current)
		return nil
	}

	asset := assetName(binary)
	base := fmt.Sprintf("https://github.com/%s/releases/download/%s", repo(), tag)

	fmt.Printf("updating heimdall-%s: %s -> %s\n", binary, current, tag)
	bin, err := get(base + "/" + asset)
	if err != nil {
		return fmt.Errorf("download %s: %w", asset, err)
	}
	manifest, err := get(base + "/SHA256SUMS")
	if err != nil {
		return fmt.Errorf("download checksums: %w", err)
	}
	want, ok := sumFor(string(manifest), asset)
	if !ok {
		return fmt.Errorf("no checksum for %s in release %s", asset, tag)
	}
	got := hex.EncodeToString(sha256Sum(bin))
	if got != want {
		return fmt.Errorf("checksum mismatch for %s: got %s want %s", asset, got, want)
	}
	if err := replaceSelf(bin); err != nil {
		return err
	}
	fmt.Printf("heimdall-%s updated to %s\n", binary, tag)
	return nil
}

func latestTag() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo())
	body, err := get(url)
	if err != nil {
		return "", err
	}
	var rel struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(body, &rel); err != nil {
		return "", err
	}
	if rel.TagName == "" {
		return "", fmt.Errorf("no published release found for %s", repo())
	}
	return rel.TagName, nil
}

func get(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "heimdall-selfupdate")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func sha256Sum(b []byte) []byte {
	h := sha256.Sum256(b)
	return h[:]
}

// replaceSelf writes the new binary next to the running executable and renames
// it into place atomically. The directory must be writable (use sudo for system
// paths like /usr/local/bin).
func replaceSelf(data []byte) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	dir := filepath.Dir(exe)
	tmp, err := os.CreateTemp(dir, ".heimdall-update-*")
	if err != nil {
		return fmt.Errorf("cannot write to %s (try sudo): %w", dir, err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	tmp.Close()
	if err := os.Chmod(tmpName, 0o755); err != nil {
		os.Remove(tmpName)
		return err
	}
	// Windows cannot replace a running image directly; move it aside first.
	if runtime.GOOS == "windows" {
		_ = os.Rename(exe, exe+".old")
	}
	if err := os.Rename(tmpName, exe); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("cannot replace %s (try sudo): %w", exe, err)
	}
	return nil
}
