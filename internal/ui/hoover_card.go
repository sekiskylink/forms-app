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

// HoverCard is a transparent tappable + hoverable overlay that sits above some content.
type HoverCard struct {
	widget.BaseWidget

	content fyne.CanvasObject
	bg      *canvas.Rectangle
	overlay *canvas.Rectangle

	onTap func()
}

func NewHoverCard(content fyne.CanvasObject, onTap func()) *HoverCard {
	// v2.6.x: use theme.BackgroundColor() / theme.HoverColor()
	bg := canvas.NewRectangle(theme.BackgroundColor())
	overlay := canvas.NewRectangle(theme.HoverColor())
	overlay.Hide()

	h := &HoverCard{
		content: content,
		bg:      bg,
		overlay: overlay,
		onTap:   onTap,
	}
	h.ExtendBaseWidget(h)
	return h
}

// --- Interaction ---

// Tapped implements fyne.Tappable
func (h *HoverCard) Tapped(_ *fyne.PointEvent) {
	if h.onTap != nil {
		h.onTap()
	}
}

// MouseIn implements desktop.Hoverable (desktop only; won't fire on mobile)
func (h *HoverCard) MouseIn(_ *desktop.MouseEvent) {
	h.overlay.Show()
	canvas.Refresh(h.overlay)
}

// MouseOut implements desktop.Hoverable
func (h *HoverCard) MouseOut() {
	h.overlay.Hide()
	canvas.Refresh(h.overlay)
}

// MouseMoved implements desktop.Hoverable
func (h *HoverCard) MouseMoved(_ *desktop.MouseEvent) {}

// --- Renderer ---

type hoverCardRenderer struct {
	card    *HoverCard
	stack   *fyne.Container // <- use *fyne.Container, not *container.Stack
	objects []fyne.CanvasObject
}

func (h *HoverCard) CreateRenderer() fyne.WidgetRenderer {
	// Stack order: bg -> content -> overlay
	st := container.NewStack(h.bg, h.content, h.overlay)
	r := &hoverCardRenderer{
		card:    h,
		stack:   st,
		objects: []fyne.CanvasObject{st},
	}
	return r
}

func (r *hoverCardRenderer) Layout(size fyne.Size) {
	r.stack.Resize(size) // ok on v2.6.x because *fyne.Container implements CanvasObject
}

func (r *hoverCardRenderer) MinSize() fyne.Size {
	min := r.card.content.MinSize()
	padding := fyne.NewSize(12, 10)
	return min.Add(padding)
}

func (r *hoverCardRenderer) Refresh() {
	// keep colors in sync with theme (v2.6.x API)
	r.card.bg.FillColor = theme.BackgroundColor()
	r.card.overlay.FillColor = theme.HoverColor()
	canvas.Refresh(r.stack)
}

func (r *hoverCardRenderer) BackgroundColor() color.Color {
	return theme.BackgroundColor()
}

func (r *hoverCardRenderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *hoverCardRenderer) Destroy()                     {}
