// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package options

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"golang.org/x/term"
)

// Locate resolves the JSON config path for a binary, e.g. <dir>/daemon.json.
func Locate(binary string) (string, error) {
	dir, err := directory()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, binary+".json"), nil
}

// directory resolves the Heimdall config directory across platforms:
//
//	$HEIMDALL_CONFIG_DIR        explicit override
//	$XDG_CONFIG_HOME/heimdall   when set
//	~/.config/heimdall          Linux & macOS (discoverable for a terminal tool)
//	%AppData%\heimdall          Windows (os.UserConfigDir)
func directory() (string, error) {
	if v := os.Getenv("HEIMDALL_CONFIG_DIR"); v != "" {
		return v, nil
	}
	if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
		return filepath.Join(v, "heimdall"), nil
	}
	if runtime.GOOS != "windows" {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			return filepath.Join(home, ".config", "heimdall"), nil
		}
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "heimdall"), nil
}

// FromFile adapts a JSON config file to a Source. The bool reports whether the
// file existed; a missing file is not an error so callers fall back to defaults.
func FromFile(path string) (Source, bool, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return mapSource{values: map[Key]string{}}, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, true, err
	}
	values := make(map[Key]string, len(raw))
	for name, value := range raw {
		if text, ok := stringify(value); ok {
			values[Key(name)] = text
		}
	}
	return mapSource{values: values}, true, nil
}

func stringify(value any) (string, bool) {
	switch typed := value.(type) {
	case string:
		return typed, true
	case bool:
		return strconv.FormatBool(typed), true
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64), true
	default:
		return "", false
	}
}

// Sink persists resolved settings as indented JSON (0600, owner-only because the
// file may hold tokens), typing each value by its option Kind.
type Sink struct{ Path string }

func (s Sink) Write(catalog Catalog, resolved Resolved) error {
	if s.Path == "" {
		return errors.New("options: empty config path")
	}
	document := make(map[string]any, catalog.Len())
	catalog.Each(func(o Option) {
		document[string(o.Key())] = typed(o, resolved.Raw(o.Key()))
	})
	body, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(s.Path, append(body, '\n'), 0o600)
}

func typed(o Option, raw string) any {
	if o.Kind() == KindToggle {
		return raw == "true"
	}
	return raw
}

// Interactive reports whether stdin is a terminal — i.e. whether it is safe to
// run the wizard (never prompt under pipes, CI, or --snapshot).
func Interactive() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}
