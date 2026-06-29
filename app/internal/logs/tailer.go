// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package logs

import (
	"bufio"
	"context"
	"io"
	"os"
	"strings"
	"time"
)

// Tail follows a file from its current end, invoking emit for each new complete
// line until ctx is cancelled (Heimdallr's sight push tail). It polls, so it
// behaves the same on macOS, Linux, and Windows. Only lines appended after the
// tail starts are emitted.
func Tail(ctx context.Context, path string, emit func(string)) error {
	return tailFile(ctx, path, emit)
}

// tailFile follows a file from its current end, invoking emit for each new
// complete line until ctx is cancelled. It polls rather than using a
// platform-specific notification API, so it behaves the same on macOS, Linux,
// and Windows. Only lines appended after the tail starts are emitted.
func tailFile(ctx context.Context, path string, emit func(string)) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	reader := bufio.NewReader(f)
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()

	var partial strings.Builder
	for {
		for {
			chunk, err := reader.ReadString('\n')
			if len(chunk) > 0 {
				partial.WriteString(chunk)
				if strings.HasSuffix(chunk, "\n") {
					emit(strings.TrimRight(partial.String(), "\r\n"))
					partial.Reset()
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
