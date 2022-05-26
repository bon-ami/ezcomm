package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type Entry struct {
	widget.Entry
	popUp *widget.PopUpMenu
}

func NewMultiLineEntry() *Entry {
	//e := &widget.Entry{MultiLine: true, Wrapping: fyne.TextTruncate}
	ret := &Entry{widget.Entry{MultiLine: true, Wrapping: fyne.TextTruncate}, nil}
	ret.ExtendBaseWidget(ret)
	return ret
}

func (e *Entry) Hide() {
	if e.popUp != nil {
		e.popUp.Hide()
		e.popUp = nil
	}
	e.Entry.Hide()
}

// selectAll cannot work
func (e *Entry) selectAll() {
	//if e.textProvider().len() == 0 {
	if e.Text == "" {
		return
	}
	/*e.setFieldsAndRefresh(func() {
		e.selectRow = 0
		e.selectColumn = 0

		lastRow := e.textProvider().rows() - 1
		e.CursorColumn = e.textProvider().rowLength(lastRow)
		e.CursorRow = lastRow
		e.selecting = true
	})*/
	e.Refresh()
}

func (e *Entry) clearAll() {
	e.SetText("")
	e.Refresh()
}

func (e *Entry) copyAll() {
	clipboard := fyne.CurrentApp().Driver().AllWindows()[0].Clipboard()
	clipboard.SetContent(e.Text)
}

func (e *Entry) TappedSecondary(pe *fyne.PointEvent) {
	if e.Disabled() && e.Password {
		return // no popup options for a disabled concealed field
	}
	//e.requestFocus()
	fyne.CurrentApp().Driver().CanvasForObject(e).Focus(e)
	clipboard := fyne.CurrentApp().Driver().AllWindows()[0].Clipboard()
	//super := e.super()

	cutItem := fyne.NewMenuItem("Cut", func() {
		e.Entry.TypedShortcut(&fyne.ShortcutCut{Clipboard: clipboard})
		//super.(fyne.Shortcutable).TypedShortcut(&fyne.ShortcutCut{Clipboard: clipboard})
	})
	copyItem := fyne.NewMenuItem("Copy", func() {
		e.Entry.TypedShortcut(&fyne.ShortcutCopy{Clipboard: clipboard})
	})
	pasteItem := fyne.NewMenuItem("Paste", func() {
		e.Entry.TypedShortcut(&fyne.ShortcutPaste{Clipboard: clipboard})
	})
	//selectAllItem := fyne.NewMenuItem("Select all", e.selectAll)
	copyAllItem := fyne.NewMenuItem("Copy all", e.copyAll)
	clearItem := fyne.NewMenuItem("Clear", e.clearAll)

	entryPos := fyne.CurrentApp().Driver().AbsolutePositionForObject(e /*super*/)
	popUpPos := entryPos.Add(fyne.NewPos(pe.Position.X, pe.Position.Y))
	c := fyne.CurrentApp().Driver().CanvasForObject(e /*super*/)

	var menu *fyne.Menu
	if e.Disabled() {
		menu = fyne.NewMenu("", copyItem /*,selectAllItem*/, copyAllItem, clearItem)
	} else if e.Password {
		menu = fyne.NewMenu("", pasteItem /*,selectAllItem*/, copyAllItem, clearItem)
	} else {
		menu = fyne.NewMenu("", cutItem, copyItem, pasteItem /*,selectAllItem*/, copyAllItem, clearItem)
	}

	e.popUp = widget.NewPopUpMenu(menu, c)
	e.popUp.ShowAtPosition(popUpPos)
}
