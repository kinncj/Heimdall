// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package options

import "time"

// Resolver folds sources in precedence order — the last source that has a key
// wins — over a catalog, producing a Resolved snapshot.
type Resolver struct {
	catalog Catalog
	sources []Source
}

// NewResolver starts a resolver over a catalog with no sources.
func NewResolver(catalog Catalog) Resolver { return Resolver{catalog: catalog} }

// With appends a higher-precedence source, returning a new Resolver (the receiver
// is unchanged).
func (r Resolver) With(source Source) Resolver {
	next := make([]Source, len(r.sources), len(r.sources)+1)
	copy(next, r.sources)
	return Resolver{catalog: r.catalog, sources: append(next, source)}
}

// Resolve computes the effective value of every option.
func (r Resolver) Resolve() Resolved {
	values := make(map[Key]string)
	r.catalog.Each(func(o Option) { values[o.Key()] = r.pick(o.Key(), o.Fallback()) })
	return Resolved{values: values}
}

func (r Resolver) pick(key Key, fallback string) string {
	chosen := fallback
	for _, source := range r.sources {
		if v, ok := source.Value(key); ok {
			chosen = v
		}
	}
	return chosen
}

// Resolved is the read side: typed accessors over the effective values.
type Resolved struct{ values map[Key]string }

// Raw returns the effective value as a string.
func (r Resolved) Raw(key Key) string { return r.values[key] }

// Text returns a string option.
func (r Resolved) Text(key Key) string { return r.values[key] }

// Address returns an option as an Address value object.
func (r Resolved) Address(key Key) Address { return NewAddress(r.values[key]) }

// Secret returns an option as a redacting Secret value object.
func (r Resolved) Secret(key Key) Secret { return NewSecret(r.values[key]) }

// Toggle returns a bool option.
func (r Resolved) Toggle(key Key) bool { return r.values[key] == "true" }

// Span returns a duration option, falling back when unparsable.
func (r Resolved) Span(key Key, fallback time.Duration) time.Duration {
	return spanOr(r.values[key], fallback)
}
