// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package options

import (
	"flag"
	"os"
)

// Source supplies a value for a key when it has one. It is the single seam every
// input mechanism adapts to: defaults, file, environment, flags, and wizard.
type Source interface {
	Value(key Key) (string, bool)
}

// mapSource is a Source backed by a resolved key→value map.
type mapSource struct{ values map[Key]string }

func (m mapSource) Value(key Key) (string, bool) {
	v, ok := m.values[key]
	return v, ok
}

// Builtins is the lowest-precedence source: each option's declared default.
func Builtins(catalog Catalog) Source {
	values := make(map[Key]string)
	catalog.Each(func(o Option) { values[o.Key()] = o.Fallback() })
	return mapSource{values: values}
}

// FromEnvironment adapts the process environment: an option contributes when its
// EnvVar is set and non-empty.
func FromEnvironment(catalog Catalog) Source {
	values := make(map[Key]string)
	catalog.Each(func(o Option) {
		if o.EnvVar() == "" {
			return
		}
		if v, ok := os.LookupEnv(o.EnvVar()); ok && v != "" {
			values[o.Key()] = v
		}
	})
	return mapSource{values: values}
}

// FromFlags adapts the flag set: only flags the user explicitly set contribute,
// so unset flags fall through to the lower-precedence sources.
func FromFlags(catalog Catalog, set *flag.FlagSet) Source {
	provided := make(map[string]string)
	set.Visit(func(f *flag.Flag) { provided[f.Name] = f.Value.String() })
	values := make(map[Key]string)
	catalog.Each(func(o Option) {
		if v, ok := provided[o.FlagName()]; ok {
			values[o.Key()] = v
		}
	})
	return mapSource{values: values}
}

// AnyProvided reports whether the user set at least one flag.
func AnyProvided(set *flag.FlagSet) bool { return set.NFlag() > 0 }

// NoSaveRequested reports whether the user asked for a one-off run that must not
// be persisted to the config file (--no-save or its --ephemeral alias).
func NoSaveRequested(set *flag.FlagSet) bool {
	requested := false
	set.Visit(func(f *flag.Flag) {
		if f.Name == "no-save" || f.Name == "ephemeral" {
			requested = true
		}
	})
	return requested
}

// Provided reports whether the user explicitly set any flag backed by the
// catalog — i.e. a real setting, ignoring action flags like --once or --version.
func Provided(catalog Catalog, set *flag.FlagSet) bool {
	provided := make(map[string]bool)
	set.Visit(func(f *flag.Flag) { provided[f.Name] = true })
	got := false
	catalog.Each(func(o Option) {
		if provided[o.FlagName()] {
			got = true
		}
	})
	return got
}
