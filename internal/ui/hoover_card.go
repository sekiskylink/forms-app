package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func NewHoverCard(content fyne.CanvasObject, onTap func()) fyne.CanvasObject {
	// Base background that adapts to theme
	base := theme.Color(theme.ColorNameInputBackground) // theme-safe surface color
	bg := canvas.NewRectangle(base)
	bg.CornerRadius = 8

	// Hover color = subtle lighten/darken relative to base
	hover := shiftRGBA(base, 8) // small delta; positive brightens, negative darkens

	card := container.NewStack(bg, content)

	t := &hoverCard{
		Card:  card,
		Bg:    bg,
		Base:  base,
		Hover: hover,
		OnTap: onTap,
	}
	t.ExtendBaseWidget(t)
	return t
}

type hoverCard struct {
	widget.BaseWidget
	Card        *fyne.Container
	Bg          *canvas.Rectangle
	Base, Hover color.Color
	OnTap       func()
}

func (h *hoverCard) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(h.Card)
}

func (h *hoverCard) MouseIn(_ *desktop.MouseEvent) {
	h.Bg.FillColor = h.Hover
	h.Bg.Refresh()
}

func (h *hoverCard) MouseMoved(_ *desktop.MouseEvent) {}

func (h *hoverCard) MouseOut() {
	h.Bg.FillColor = h.Base
	h.Bg.Refresh()
}

func (h *hoverCard) Tapped(_ *fyne.PointEvent) {
	if h.OnTap != nil {
		h.OnTap()
	}
}

// shiftRGBA slightly brightens (>0) or darkens (<0) a theme color.
func shiftRGBA(c color.Color, delta int) color.Color {
	r, g, b, a := c.RGBA()
	rr := clamp(int(r>>8)+delta, 0, 255)
	gg := clamp(int(g>>8)+delta, 0, 255)
	bb := clamp(int(b>>8)+delta, 0, 255)
	return color.NRGBA{uint8(rr), uint8(gg), uint8(bb), uint8(a >> 8)}
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
