package ui

import (
	_ "embed"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"image/color"
)

//go:embed assets/fonts/Poppins-Regular.ttf
var poppinsRegular []byte

//go:embed assets/fonts/Poppins-Bold.ttf
var poppinsBold []byte

//go:embed assets/fonts/Poppins-Italic.ttf
var poppinsItalic []byte

// CustomTheme override fonts or colors
type CustomTheme struct {
	fyne.Theme
	variant fyne.ThemeVariant
}

// ----------------------
// Define custom colors
// ----------------------
var (
	// accentBlue   = color.NRGBA{R: 0, G: 150, B: 136, A: 255} // Teal-blue accent
	accentBlue    = color.NRGBA{R: 33, G: 150, B: 243, A: 255} // Primary blue
	accentHover   = color.NRGBA{R: 0, G: 220, B: 255, A: 255}  // lighter hover
	accentPressed = color.NRGBA{R: 25, G: 118, B: 210, A: 255} // darker press
	textOnBlue    = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	lightGray     = color.NRGBA{R: 245, G: 245, B: 245, A: 255}
	textDark      = color.NRGBA{R: 40, G: 40, B: 40, A: 255}
	textLight     = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	borderGray    = color.NRGBA{R: 220, G: 220, B: 220, A: 255}
	buttonHover   = color.NRGBA{R: 0, G: 180, B: 160, A: 255}
	errorRed      = color.NRGBA{R: 220, G: 0, B: 0, A: 255}
	successGreen  = color.NRGBA{R: 30, G: 180, B: 90, A: 255}
)

func (c *CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		if variant == theme.VariantDark {
			return color.NRGBA{R: 18, G: 18, B: 18, A: 255}
		}
		return lightGray

	case theme.ColorNameButton:
		return accentBlue

	case theme.ColorNameForeground:
		if variant == theme.VariantDark {
			return textLight
		}
		return textDark

	case theme.ColorNameHover:
		return buttonHover

	case theme.ColorNamePrimary:
		return accentBlue

	case theme.ColorNameError:
		return errorRed

	case theme.ColorNameSuccess:
		return successGreen

	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}

	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (CustomTheme) Font(style fyne.TextStyle) fyne.Resource {
	switch {
	case style.Bold:
		return fyne.NewStaticResource("Poppins-Bold.ttf", poppinsBold)
	case style.Italic:
		return fyne.NewStaticResource("Poppins-Italic.ttf", poppinsItalic)
	default:
		return fyne.NewStaticResource("Poppins-Regular.ttf", poppinsRegular)
	}
	// return theme.DefaultTheme().Font(style)
}

func (CustomTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (c *CustomTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		// ⬆️ make all normal text (including Entry text) larger
		return theme.DefaultTheme().Size(name) * 1.1
	case theme.SizeNameHeadingText:
		return theme.DefaultTheme().Size(name) * 1.1
	case theme.SizeNamePadding:
		return 8
	default:
		return theme.DefaultTheme().Size(name)
	}
}

// ApplyCustomTheme sets this theme as active.
// CustomTheme wraps fyne.Theme to override the default font.
func ApplyCustomTheme(a fyne.App) {
	a.Settings().SetTheme(&CustomTheme{Theme: theme.DefaultTheme()})
}
