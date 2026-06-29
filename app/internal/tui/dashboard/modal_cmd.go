// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package dashboard

import (
	"fmt"
	"strings"

	"heimdall/app/internal/command"
	"heimdall/app/internal/domain"
)

// labelCmd marks a host that exposes on-demand commands (daemon --allow-commands).
const labelCmd = "_cmd"

// hasCmd reports whether the host exposes on-demand commands (gates the modal).
func hasCmd(h domain.HostView) bool { return h.Host.Context.Labels[labelCmd] != "" }

// cmdModalKeys are the no-argument allow-listed commands offered in the modal.
// dir.list needs a path, so it stays CLI-only for now.
func cmdModalKeys() []string {
	out := make([]string, 0, len(command.Keys()))
	for _, k := range command.Keys() {
		if k != "dir.list" {
			out = append(out, k)
		}
	}
	return out
}

func (m Model) cmdResultName() string {
	keys := cmdModalKeys()
	if m.cmdSel < len(keys) {
		return keys[m.cmdSel]
	}
	return ""
}

func (m Model) cmdListBody() []string {
	val, _ := m.mode.Role("value")
	focus, _ := m.mode.Role("focus")
	keys := cmdModalKeys()
	out := make([]string, len(keys))
	for i, k := range keys {
		if i == m.cmdSel {
			out[i] = focus.Style().Render("  ▸ " + k)
		} else {
			out[i] = val.Style().Render("    " + k)
		}
	}
	return out
}

func (m Model) cmdResultBody(h domain.HostView, w int) []string {
	muted, _ := m.mode.Role("text_muted")
	val, _ := m.mode.Role("value")
	if m.runCmd == nil {
		return []string{muted.Style().Render("  commands are unavailable (no hub connection)")}
	}
	cr := h.LastCommand
	if cr == nil || cr.RequestID != m.cmdReqID {
		return []string{muted.Style().Render("  running…")}
	}

	head := fmt.Sprintf("  status: %s   exit: %d", statusWord(cr.Status), cr.ExitCode)
	if cr.Status != domain.StatusOK {
		st, _ := m.mode.State("error")
		head = st.Style().Render(head)
	} else {
		head = muted.Style().Render(head)
	}
	out := []string{head, ""}

	text := cr.Stdout
	if strings.TrimSpace(cr.Stderr) != "" {
		text += cr.Stderr
	}
	for _, line := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		out = append(out, "  "+val.Style().Render(clip(line, w-4)))
	}
	if cr.Truncated {
		out = append(out, muted.Style().Render("  [output truncated]"))
	}
	return out
}

func statusWord(s domain.MetricStatus) string {
	switch s {
	case domain.StatusOK:
		return "ok"
	case domain.StatusInsufficientPermission:
		return "insufficient_permission"
	case domain.StatusError:
		return "error"
	default:
		return "unspecified"
	}
}
