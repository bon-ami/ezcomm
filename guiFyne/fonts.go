package guiFyne

import "fyne.io/fyne/v2"

var FontsBuiltin = [...]struct {
	locale string
	res    *fyne.StaticResource
}{
	{"zh_CN", FontZhCn},
}
