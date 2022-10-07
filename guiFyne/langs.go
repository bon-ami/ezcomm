package main

import "fyne.io/fyne/v2"

var LangsBuiltin = [...]struct {
	locale         string
	fnt, name, rbt *fyne.StaticResource
}{
	{"en", nil, nil, nil},
	{"es", nil, nil, nil},
	{"ja", FontJa, nil, nil},
	{"zh-CN", FontZhCn, nil, nil},
	{"zh-TW", FontZhTw, nil, nil},
}
