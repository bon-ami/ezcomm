package main

import "fyne.io/fyne/v2"

var LangsBuiltin = [...]struct {
	locale                                     string
	fnt, name, rbt                             *fyne.StaticResource
	nameWidth, nameHeight, rbtWidth, rbtHeight float32
}{
	{"en", nil, IconLangEn, nil, 53, 14, -1, -1},
	{"es", nil, IconLangEs, nil, 53, 16, -1, -1},
	{"ja", FontJa, IconLangJa, IconRbtJa, 44, 16, 203, 60},
	{"zh-CN", FontZhCn, IconLangZhCn, IconRbtZhCn, 28, 15, 231, 20},
	{"zh-TW", FontZhTw, IconLangZhTw, IconRbtZhTw, 60, 14, 200, 20},
}
