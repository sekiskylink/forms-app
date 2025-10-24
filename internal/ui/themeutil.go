package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// variantTheme forces a given theme variant (light/dark) but otherwise uses the default theme.
type variantTheme struct {
	base    fyne.Theme
	variant fyne.ThemeVariant
}

func (t variantTheme) Color(n fyne.ThemeColorName, _ fyne.ThemeVariant) (c color.Color) {
	return t.base.Color(n, t.variant)
}
func (t variantTheme) Font(s fyne.TextStyle) fyne.Resource     { return t.base.Font(s) }
func (t variantTheme) Icon(n fyne.ThemeIconName) fyne.Resource { return t.base.Icon(n) }
func (t variantTheme) Size(n fyne.ThemeSizeName) float32       { return t.base.Size(n) }

// SetDark toggles using the wrapper.
func SetDark(a fyne.App, dark bool) {
	v := theme.VariantLight
	if dark {
		v = theme.VariantDark
	}
	a.Settings().SetTheme(variantTheme{base: theme.DefaultTheme(), variant: v})
}
