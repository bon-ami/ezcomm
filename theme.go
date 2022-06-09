package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"gitee.com/bon-ami/eztools/v4"
)

type theme4Fonts struct {
	fontRes fyne.Resource
}

func (m *theme4Fonts) SetFont(font string) error {
	if len(font) < 1 {
		return eztools.ErrInvalidInput
	}
	var err error
	m.fontRes, err = fyne.LoadResourceFromPath(font)
	return err
}

func (m theme4Fonts) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameBackground {
		if variant == theme.VariantLight {
			return color.White
		}
		return color.Black
	}

	return theme.DefaultTheme().Color(name, variant)
}

func (m theme4Fonts) Icon(name fyne.ThemeIconName) fyne.Resource {
	/*if name == theme.IconNameHome {
		fyne.NewStaticResource("myHome", homeBytes)
	}*/

	return theme.DefaultTheme().Icon(name)
}
func (m theme4Fonts) Font(style fyne.TextStyle) fyne.Resource {
	if m.fontRes != nil {
		return m.fontRes
	}
	return theme.DefaultTheme().Font(style)
}

func (m theme4Fonts) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
