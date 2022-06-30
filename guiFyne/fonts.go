package main

import "fyne.io/fyne/v2"

var FontsBuiltin = [...]struct {
	locale string
	res    *fyne.StaticResource
}{
	{"zh-CN", FontZhCn},
	{"zh-TW", FontZhTw},
}
