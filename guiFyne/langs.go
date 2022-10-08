package main

import "fyne.io/fyne/v2"

var LangsBuiltin = [...]struct {
	locale         string
	fnt, name, rbt *fyne.StaticResource
	width, height  float32
}{
	{"en", nil, IconLangEn, nil, 53, 14},
	{"es", nil, IconLangEs, nil, 53, 16},
	{"ja", FontJa, IconLangJa, nil, 44, 16},
	{"zh-CN", FontZhCn, IconLangZhCn, nil, 28, 15},
	{"zh-TW", FontZhTw, IconLangZhTw, nil, 60, 14},
}
