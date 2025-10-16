package ui

import (
	"fyne.io/fyne/v2"
)

// Navigator manages screen transitions and history.
type Navigator struct {
	window fyne.Window
	stack  []fyne.CanvasObject
}

// NewNavigator creates a navigation manager bound to a window.
func NewNavigator(w fyne.Window) *Navigator {
	return &Navigator{
		window: w,
		stack:  make([]fyne.CanvasObject, 0),
	}
}

// Reset clears the navigation stack and sets a new root instantly.
func (n *Navigator) Reset(screen fyne.CanvasObject) {
	n.stack = []fyne.CanvasObject{screen}
	n.window.SetContent(screen)
}

// Current returns the current screen on top of the stack.
func (n *Navigator) Current() fyne.CanvasObject {
	if len(n.stack) == 0 {
		return nil
	}
	return n.stack[len(n.stack)-1]
}

//
// ─── SLIDE ANIMATIONS ───────────────────────────────────────────────
//

// PushSlide navigates forward to a new screen (right → left).
func (n *Navigator) PushSlide(screen fyne.CanvasObject) {
	n.stack = append(n.stack, screen)
	AnimateSlideIn(n.window, screen)
}

// PopSlide navigates back (left → right).
func (n *Navigator) PopSlide() {
	if len(n.stack) <= 1 {
		return
	}
	n.stack = n.stack[:len(n.stack)-1]
	prev := n.stack[len(n.stack)-1]
	AnimateSlideOut(n.window, prev)
}

//
// ─── FADE ANIMATIONS ────────────────────────────────────────────────
//

// PushFade navigates forward with a fade-in.
func (n *Navigator) PushFade(screen fyne.CanvasObject) {
	n.stack = append(n.stack, screen)
	AnimateFadeIn(n.window, screen)
}

// PopFade fades back to the previous screen.
func (n *Navigator) PopFade() {
	if len(n.stack) <= 1 {
		return
	}
	n.stack = n.stack[:len(n.stack)-1]
	prev := n.stack[len(n.stack)-1]
	AnimateFadeIn(n.window, prev)
}
