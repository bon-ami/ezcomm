package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"gitee.com/bon-ami/eztools/v6"
)

type theme4Fonts struct {
	fontRes fyne.Resource
}

func (m *theme4Fonts) SetFontByRes(font fyne.Resource) {
	m.fontRes = font
}

func (m *theme4Fonts) SetFontByDir(font string) error {
	if len(font) < 1 {
		return eztools.ErrInvalidInput
	}
	var err error
	m.fontRes, err = fyne.LoadResourceFromPath(font)
	return err
}

func (m theme4Fonts) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(name, variant)
}

func (m theme4Fonts) Icon(name fyne.ThemeIconName) fyne.Resource {
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
