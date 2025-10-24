package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop" // ✅ required for MouseIn/Out/Down/Up
	"fyne.io/fyne/v2/widget"
)

type AccentButton struct {
	widget.BaseWidget
	Text      string
	OnTapped  func()
	hovered   bool
	pressed   bool
	bgColor   color.Color
	textColor color.Color
}

func NewAccentButton(text string, onTapped func()) *AccentButton {
	btn := &AccentButton{
		Text:      text,
		OnTapped:  onTapped,
		bgColor:   accentBlue,
		textColor: textOnBlue,
	}
	btn.ExtendBaseWidget(btn)
	return btn
}

// --- Mouse interactivity ---
func (b *AccentButton) MouseIn(*desktop.MouseEvent)   { b.hovered = true; b.Refresh() }
func (b *AccentButton) MouseOut()                     { b.hovered = false; b.Refresh() }
func (b *AccentButton) MouseDown(*desktop.MouseEvent) { b.pressed = true; b.Refresh() }
func (b *AccentButton) MouseUp(*desktop.MouseEvent) {
	b.pressed = false
	if b.OnTapped != nil {
		b.OnTapped()
	}
	b.Refresh()
}
func (b *AccentButton) MouseMoved(*desktop.MouseEvent) {}

func (b *AccentButton) Tapped(_ *fyne.PointEvent) {
	if b.OnTapped != nil {
		b.OnTapped()
	}
}

// --- Renderer ---
func (b *AccentButton) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(b.bgColor)
	bg.CornerRadius = 10 // ✅ Rounded corners

	label := canvas.NewText(b.Text, b.textColor)
	label.Alignment = fyne.TextAlignCenter
	label.TextSize = 16

	shadow := canvas.NewRectangle(color.NRGBA{0, 0, 0, 25}) // soft drop shadow
	shadow.CornerRadius = 10

	objects := []fyne.CanvasObject{shadow, bg, label}
	return &accentButtonRenderer{btn: b, bg: bg, label: label, shadow: shadow, objects: objects}
}

type accentButtonRenderer struct {
	btn     *AccentButton
	bg      *canvas.Rectangle
	label   *canvas.Text
	shadow  *canvas.Rectangle
	objects []fyne.CanvasObject
}

func (r *accentButtonRenderer) Layout(size fyne.Size) {
	// Drop shadow slightly offset
	r.shadow.Move(fyne.NewPos(0, 3))
	r.shadow.Resize(size)

	r.bg.Resize(size)
	r.bg.Move(fyne.NewPos(0, 0))

	r.label.Resize(size)
	r.label.Move(fyne.NewPos(0, (size.Height-r.label.MinSize().Height)/2))
}

func (r *accentButtonRenderer) MinSize() fyne.Size {
	return fyne.NewSize(140, 46)
}

func (r *accentButtonRenderer) Refresh() {
	var col color.Color

	switch {
	case r.btn.pressed:
		col = color.NRGBA{R: 13, G: 71, B: 161, A: 255} // pressed
	case r.btn.hovered:
		col = color.NRGBA{R: 0, G: 210, B: 255, A: 255} // vivid hover cyan
	default:
		col = color.NRGBA{R: 33, G: 150, B: 243, A: 255} // normal blue
	}

	r.bg.FillColor = col
	canvas.Refresh(r.bg)
	canvas.Refresh(r.label)
}

func (r *accentButtonRenderer) Destroy()                     {}
func (r *accentButtonRenderer) Objects() []fyne.CanvasObject { return r.objects }

type SecondaryButton struct {
	widget.BaseWidget
	Text      string
	OnTapped  func()
	hovered   bool
	pressed   bool
	borderCol color.Color
	textColor color.Color
}

func NewSecondaryButton(text string, onTapped func()) *SecondaryButton {
	btn := &SecondaryButton{
		Text:      text,
		OnTapped:  onTapped,
		borderCol: accentBlue,
		textColor: accentBlue,
	}
	btn.ExtendBaseWidget(btn)
	return btn
}

func (b *SecondaryButton) MouseIn(*desktop.MouseEvent)   { b.hovered = true; b.Refresh() }
func (b *SecondaryButton) MouseOut()                     { b.hovered = false; b.Refresh() }
func (b *SecondaryButton) MouseDown(*desktop.MouseEvent) { b.pressed = true; b.Refresh() }
func (b *SecondaryButton) MouseUp(*desktop.MouseEvent) {
	b.pressed = false
	if b.OnTapped != nil {
		b.OnTapped()
	}
	b.Refresh()
}

func (b *SecondaryButton) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(color.NRGBA{255, 255, 255, 255})
	bg.CornerRadius = 10
	border := canvas.NewRectangle(b.borderCol)
	border.CornerRadius = 10

	label := canvas.NewText(b.Text, b.textColor)
	label.Alignment = fyne.TextAlignCenter
	label.TextSize = 16

	objects := []fyne.CanvasObject{border, bg, label}
	return &secondaryButtonRenderer{
		btn:     b,
		bg:      bg,
		border:  border,
		label:   label,
		objects: objects,
	}
}

type secondaryButtonRenderer struct {
	btn     *SecondaryButton
	bg      *canvas.Rectangle
	border  *canvas.Rectangle
	label   *canvas.Text
	objects []fyne.CanvasObject
}

func (r *secondaryButtonRenderer) Layout(size fyne.Size) {
	r.border.Resize(size)
	r.bg.Resize(fyne.NewSize(size.Width-2, size.Height-2))
	r.bg.Move(fyne.NewPos(1, 1))
	r.label.Resize(size)
	r.label.Move(fyne.NewPos(0, (size.Height-r.label.MinSize().Height)/2))
}

func (r *secondaryButtonRenderer) MinSize() fyne.Size {
	return fyne.NewSize(140, 46)
}

func (r *secondaryButtonRenderer) Refresh() {
	bgCol := color.NRGBA{255, 255, 255, 255} // default white
	borderCol := accentBlue
	textCol := accentBlue

	if r.btn.hovered {
		bgCol = color.NRGBA{230, 240, 255, 255} // soft blue tint
	}
	if r.btn.pressed {
		bgCol = color.NRGBA{220, 230, 250, 255} // slightly deeper on press
	}

	r.bg.FillColor = bgCol
	r.border.FillColor = borderCol
	r.label.Color = textCol

	canvas.Refresh(r.bg)
	canvas.Refresh(r.label)
}

func (r *secondaryButtonRenderer) Destroy()                     {}
func (r *secondaryButtonRenderer) Objects() []fyne.CanvasObject { return r.objects }
