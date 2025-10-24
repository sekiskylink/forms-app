package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// tappableOverlay is a transparent layer that can detect taps (used to close drawers)
type tappableOverlay struct {
	widget.BaseWidget
	OnTapped func()
}

func newTappableOverlay(onTapped func()) *tappableOverlay {
	t := &tappableOverlay{OnTapped: onTapped}
	t.ExtendBaseWidget(t)
	return t
}

func (t *tappableOverlay) CreateRenderer() fyne.WidgetRenderer {
	rect := canvas.NewRectangle(color.NRGBA{0, 0, 0, 80}) // semi-transparent black
	return widget.NewSimpleRenderer(rect)
}

func (t *tappableOverlay) Tapped(_ *fyne.PointEvent) {
	if t.OnTapped != nil {
		t.OnTapped()
	}
}
