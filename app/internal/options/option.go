// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package options resolves per-binary settings from layered sources — built-in
// defaults, a JSON config file, the environment, and command-line flags — and
// drives a first-run setup wizard. Each source is an adapter behind one Source
// interface; the Resolver composes them by precedence (flag > flag-set config >
// env > default). Adding a setting is a single Option entry; adding an input
// mechanism is a single Source.
package options

import (
	"flag"
	"time"
)

// Key identifies an option across every source and is also its JSON field and
// (by default) its flag name.
type Key string

// Kind tells the catalog how to register a flag, the sink how to type the JSON
// value, and the resolver how to read it back.
type Kind int

const (
	KindText   Kind = iota // string
	KindToggle             // bool
	KindSpan               // time.Duration ("2s")
	KindSecret             // string, redacted when displayed
)

// Option describes one setting end to end: its key/flag/env, default, help text,
// kind, and — when set — the wizard question that makes it askable.
type Option struct {
	key      Key
	flagName string
	envVar   string
	fallback string
	help     string
	question string
	kind     Kind
	askable  bool
}

// Define starts a fluent Option whose flag name defaults to the key.
func Define(key Key) Option { return Option{key: key, flagName: string(key)} }

func (o Option) Flag(name string) Option    { o.flagName = name; return o }
func (o Option) Env(name string) Option     { o.envVar = name; return o }
func (o Option) Default(v string) Option    { o.fallback = v; return o }
func (o Option) Help(text string) Option    { o.help = text; return o }
func (o Option) Of(kind Kind) Option        { o.kind = kind; return o }
func (o Option) Ask(question string) Option { o.question, o.askable = question, true; return o }

func (o Option) Key() Key         { return o.key }
func (o Option) FlagName() string { return o.flagName }
func (o Option) EnvVar() string   { return o.envVar }
func (o Option) Fallback() string { return o.fallback }
func (o Option) Summary() string  { return o.help }
func (o Option) Question() string { return o.question }
func (o Option) Kind() Kind       { return o.kind }
func (o Option) Askable() bool    { return o.askable }

// Catalog is an immutable, ordered set of options for one binary.
type Catalog struct{ items []Option }

// NewCatalog builds a catalog from options in declaration order.
func NewCatalog(items ...Option) Catalog { return Catalog{items: items} }

// Each visits every option in order.
func (c Catalog) Each(visit func(Option)) {
	for _, o := range c.items {
		visit(o)
	}
}

// Len reports the number of options.
func (c Catalog) Len() int { return len(c.items) }

// Register declares each option as a typed flag on set, using the option's
// default for the flag default so --help shows the effective baseline.
func (c Catalog) Register(set *flag.FlagSet) {
	c.Each(func(o Option) {
		switch o.Kind() {
		case KindToggle:
			set.Bool(o.FlagName(), o.Fallback() == "true", o.Summary())
		case KindSpan:
			set.Duration(o.FlagName(), spanOr(o.Fallback(), 0), o.Summary())
		default:
			set.String(o.FlagName(), o.Fallback(), o.Summary())
		}
	})
}

func spanOr(text string, fallback time.Duration) time.Duration {
	if d, err := time.ParseDuration(text); err == nil {
		return d
	}
	return fallback
}
