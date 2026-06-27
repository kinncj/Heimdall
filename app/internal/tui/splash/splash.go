// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

// Package splash renders the Heimdall brand splash. On terminals that support an
// inline-image protocol (Kitty/iTerm/Sixel) it renders LOGO_NO_BG.png directly;
// otherwise it falls back to a clean luminance-ramp ASCII of the icon. Set
// HEIMDALL_ASCII=1 to force the ASCII form.
package splash

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png"
	"os"
	"strconv"
	"strings"

	"github.com/BourgeoisBear/rasterm"
	"github.com/charmbracelet/lipgloss"
	xdraw "golang.org/x/image/draw"

	"heimdall/app/internal/tui/theme"
)

//go:embed assets/ICON_NO_BG.png
var iconPNG []byte

//go:embed assets/LOGO_NO_BG.png
var logoPNG []byte

// ramp maps luminance (low -> high) to increasingly dense glyphs.
var ramp = []rune(" .:-=+*#%@")

func decode(b []byte) image.Image {
	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return nil
	}
	return img
}

// imageToASCII resizes img to fit cols (preserving aspect for ~2:1 cells) and
// maps each pixel to a ramp glyph; fully-transparent pixels become spaces, so a
// transparent logo renders as a clean silhouette.
func imageToASCII(img image.Image, cols int) []string {
	if img == nil {
		return nil
	}
	b := img.Bounds()
	rows := cols * b.Dy() / (b.Dx() * 2)
	if rows < 1 {
		rows = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, cols, rows))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), img, b, xdraw.Over, nil)

	lines := make([]string, rows)
	for y := 0; y < rows; y++ {
		var sb strings.Builder
		for x := 0; x < cols; x++ {
			r, g, bl, a := dst.At(x, y).RGBA()
			if a < 0x6000 {
				sb.WriteByte(' ')
				continue
			}
			lum := (0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(bl)) / 65535.0
			idx := int(lum*float64(len(ramp)-1) + 0.5)
			if idx < 0 {
				idx = 0
			} else if idx >= len(ramp) {
				idx = len(ramp) - 1
			}
			sb.WriteRune(ramp[idx])
		}
		lines[y] = strings.TrimRight(sb.String(), " ")
	}
	return lines
}

func frame(m theme.Mode, art string, width, height int) string {
	steel, _ := m.Role("title")
	accent, _ := m.Role("accent")
	block := lipgloss.JoinVertical(lipgloss.Center,
		art,
		"",
		steel.Style().Render("⬢  H E I M D A L L"),
		accent.Style().Render("— watch over all realms —"),
	)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, block)
}

// resizeImage scales img down to width w (px), preserving aspect.
func resizeImage(img image.Image, w int) image.Image {
	b := img.Bounds()
	if b.Dx() <= w {
		return img
	}
	h := w * b.Dy() / b.Dx()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), img, b, xdraw.Over, nil)
	return dst
}

func forceASCII() bool { return os.Getenv("HEIMDALL_ASCII") != "" }

// inlineImage renders LOGO_NO_BG.png with an inline-image protocol at a known
// cell footprint, centred in width x height. ok=false if the terminal can't
// display images. The logo already contains the wordmark + tagline, so nothing
// else is drawn.
func inlineImage(width, height int) (string, bool) {
	if forceASCII() || rasterm.IsTmuxScreen() {
		return "", false
	}
	img := decode(logoPNG)
	if img == nil {
		return "", false
	}
	img = resizeImage(img, 360)

	cols := 40
	if cols > width-4 {
		cols = width - 4
	}
	rows := cols / 2 // square logo, ~2:1 cell aspect
	if rows > height-4 {
		rows = height - 4
		cols = rows * 2
	}
	if cols < 8 || rows < 4 {
		return "", false
	}
	top := (height - rows) / 2
	if top < 0 {
		top = 0
	}
	left := (width - cols) / 2
	if left < 0 {
		left = 0
	}

	var buf bytes.Buffer
	buf.WriteString(strings.Repeat("\n", top)) // vertical centre
	buf.WriteString(strings.Repeat(" ", left)) // horizontal centre
	switch {
	case rasterm.IsKittyCapable():
		if rasterm.KittyWriteImage(&buf, img, rasterm.KittyImgOpts{DstCols: uint32(cols), DstRows: uint32(rows)}) != nil {
			return "", false
		}
	case rasterm.IsItermCapable():
		opts := rasterm.ItermImgOpts{DisplayInline: true, Width: strconv.Itoa(cols), Height: strconv.Itoa(rows)}
		if rasterm.ItermWriteImageWithOptions(&buf, img, opts) != nil {
			return "", false
		}
	default:
		return "", false
	}
	return buf.String(), true
}

// ASCII renders the clean luminance-ramp icon + wordmark + tagline.
func ASCII(m theme.Mode, width, height int) string {
	cols := width - 8
	if cols > 44 {
		cols = 44
	}
	if cols < 28 {
		cols = 28
	}
	steel, _ := m.Role("title")
	art := steel.Style().Render(strings.Join(imageToASCII(decode(iconPNG), cols), "\n"))
	return frame(m, art, width, height)
}

// Render returns the splash: a centred inline image where supported (the logo
// already includes the wordmark + tagline), else clean ASCII with the wordmark.
func Render(m theme.Mode, width, height int) string {
	if img, ok := inlineImage(width, height); ok {
		return img
	}
	return ASCII(m, width, height)
}
