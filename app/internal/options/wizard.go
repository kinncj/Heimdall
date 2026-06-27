// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package options

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Prompter reads the user's answers during the first-run wizard. In and Out are
// injectable so the prompts are testable without a real terminal.
type Prompter struct {
	in  *bufio.Reader
	out io.Writer
}

// NewPrompter builds a Prompter over in/out.
func NewPrompter(in io.Reader, out io.Writer) *Prompter {
	return &Prompter{in: bufio.NewReader(in), out: out}
}

// Intro prints a banner and the path the result will be written to.
func (p *Prompter) Intro(title string, lines ...string) {
	fmt.Fprintf(p.out, "\n%s\n", title)
	for _, l := range lines {
		fmt.Fprintf(p.out, "  %s\n", l)
	}
	fmt.Fprintln(p.out)
}

// Note prints an informational line.
func (p *Prompter) Note(line string) { fmt.Fprintf(p.out, "  %s\n", line) }

// Ask prompts with the current value as the default; empty input keeps it.
func (p *Prompter) Ask(label, fallback string) string {
	if fallback != "" {
		fmt.Fprintf(p.out, "  %s [%s]: ", label, fallback)
	} else {
		fmt.Fprintf(p.out, "  %s: ", label)
	}
	text, _ := p.in.ReadString('\n')
	text = strings.TrimRight(text, "\r\n")
	if strings.TrimSpace(text) == "" {
		return fallback
	}
	return text
}

// Wizard remembers the user's answers and serves them as a high-precedence
// Source.
type Wizard struct{ answers map[Key]string }

func (w Wizard) Value(key Key) (string, bool) {
	v, ok := w.answers[key]
	return v, ok
}

// RunWizard asks each askable option, showing its currently-resolved value as the
// default, and returns the answers as a Source.
func RunWizard(catalog Catalog, current Resolved, prompter *Prompter) Wizard {
	answers := make(map[Key]string)
	catalog.Each(func(o Option) {
		if !o.Askable() {
			return
		}
		answers[o.Key()] = prompter.Ask(o.Question(), current.Raw(o.Key()))
	})
	return Wizard{answers: answers}
}
