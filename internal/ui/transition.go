package ui

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// AnimateFadeIn replaces the window content with newContent
// and performs a smooth fade overlay (thread-safe via fyne.Do).
func AnimateFadeIn(w fyne.Window, newContent fyne.CanvasObject) {
	// Step 1 – set content first so it’s instantly visible
	w.SetContent(newContent)
	w.Canvas().Refresh(newContent)

	// Step 2 – create a fading overlay
	rect := canvas.NewRectangle(color.NRGBA{0, 0, 0, 200}) // semi-transparent black
	rect.Resize(w.Canvas().Size())
	overlay := container.NewStack(newContent, rect)
	w.SetContent(overlay)
	canvas.Refresh(overlay)

	// Step 3 – animate fade-out on the UI thread
	fyne.Do(func() {
		const steps = 15
		const delay = 15 * time.Millisecond

		for i := steps; i >= 0; i-- {
			alpha := uint8(float64(i) / float64(steps) * 200) // 200 → 0
			rect.FillColor = color.NRGBA{0, 0, 0, alpha}
			canvas.Refresh(rect)
			time.Sleep(delay)
		}

		// Step 4 – cleanly remove overlay
		w.SetContent(newContent)
	})
}

func AnimateSlideIn(w fyne.Window, newContent fyne.CanvasObject) {
	size := w.Canvas().Size()

	// Ensure content fills the window
	newContent.Resize(size)

	// Create a background snapshot (optional fade behind)
	bg := canvas.NewRectangle(color.NRGBA{0, 0, 0, 0})
	bg.Resize(size)

	// Place the new content initially offscreen (to the right)
	offsetX := size.Width
	pos := fyne.NewPos(offsetX, 0)
	newContent.Move(pos)

	stack := container.NewWithoutLayout(bg, newContent)
	w.SetContent(stack)

	const steps = 25
	const delay = 10 * time.Millisecond
	delta := offsetX / float32(steps)

	// Run the slide on the UI thread
	fyne.Do(func() {
		for i := 0; i <= steps; i++ {
			x := offsetX - delta*float32(i)
			newContent.Move(fyne.NewPos(x, 0))
			canvas.Refresh(stack)
			time.Sleep(delay)
		}

		// Finalize position (snap to 0,0)
		newContent.Move(fyne.NewPos(0, 0))
		canvas.Refresh(stack)

		// Replace window content with clean version (no overlay container)
		w.SetContent(newContent)
	})
}

// AnimateSlideOut slides currentContent out to the right,
// revealing newContent underneath.
func AnimateSlideOut(w fyne.Window, newContent fyne.CanvasObject) {
	size := w.Canvas().Size()
	current := w.Content() // current visible content

	// Make sure both contents are sized correctly
	current.Resize(size)
	newContent.Resize(size)

	// Stack: newContent at base, current content on top (will slide out)
	stack := container.NewWithoutLayout(newContent, current)
	w.SetContent(stack)
	canvas.Refresh(stack)

	const steps = 25
	const delay = 10 * time.Millisecond
	delta := size.Width / float32(steps)

	// Animate on UI thread
	fyne.Do(func() {
		for i := 0; i <= steps; i++ {
			x := delta * float32(i)
			current.Move(fyne.NewPos(x, 0))
			canvas.Refresh(stack)
			time.Sleep(delay)
		}

		// After animation completes, set clean new content
		w.SetContent(newContent)
	})
}
