package main

import (
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/brianvoe/gofakeit/v6"
)

/* This example show an AutoComplete Entry, just like the one in fyne-x
   but improved (simpler code, multiligne items...)
*/

func main() {
	a := app.New()
	w := a.NewWindow("AutoComplete Entry")

	var randomPersons []string
	for i := 0; i < 100; i++ {
		randomPersons = append(randomPersons, gofakeit.FirstName()+" "+gofakeit.LastName())
	}
	sort.Strings(randomPersons)

	ac := NewAutoComplete(1)
	ac.OnChanged = func(s string) {
		if s == "" {
			ac.ListHide()
			return
		}

		// you populate the list here
		// do it as you want (DB calls, Results filtering...)
		ac.Options = ac.Options[:0]
		for i := 0; i < len(randomPersons); i++ {
			if strings.HasPrefix(strings.ToLower(randomPersons[i]), strings.ToLower(s)) {
				ac.Options = append(ac.Options, randomPersons[i])
			}
		}
		ac.ListShow()
	}

	w.SetContent(container.NewBorder(ac, nil, nil, nil, widget.NewLabel("")))
	w.ShowAndRun()
}

// AutoComplete is an extended widget.Entry that can display a list of suggestions
// in a list.
//
// You control the options displayed and when they are displayed.
//
// You can also use custom CanvasObjects as list items.
//
// You can navigate through the suggested items and select them with mouse
// or keyboard (up and down arrow, enter to select, escape to hide the list).
//
// You can override what will be completed using the OnCompleted callback
// (just return something different).
type AutoComplete struct {
	widget.Entry

	// autocomplete
	Options           []string
	OnCompleted       func(string) string
	SubmitOnCompleted bool // if true, completing from list triggers OnSubmited

	CustomCreate func() fyne.CanvasObject
	CustomUpdate func(id widget.ListItemID, co fyne.CanvasObject)

	popup    *widget.PopUp
	list     *autoCompleteList
	selected widget.ListItemID
	pause    bool
}

func NewAutoComplete(minLines int) *AutoComplete {
	ac := &AutoComplete{}
	ac.ExtendBaseWidget(ac)
	if minLines > 1 {
		ac.Entry.MultiLine = true
		ac.Entry.Wrapping = fyne.TextWrapWord
		ac.Entry.SetMinRowsVisible(minLines)
	} else {
		ac.Entry.Wrapping = fyne.TextTruncate
	}
	return ac
}

func (ac *AutoComplete) AcceptsTab() bool { return ac.Entry.MultiLine }

// ---

func (ac *AutoComplete) ListShow() {
	if ac.pause {
		return
	}
	if len(ac.Options) <= 0 {
		ac.ListHide()
		return
	}

	cnv := fyne.CurrentApp().Driver().CanvasForObject(ac)
	if cnv == nil {
		return // not show
	}

	if ac.list == nil {
		ac.list = newAutoCompleteList(ac)
	}
	if ac.popup == nil {
		ac.popup = widget.NewPopUp(ac.list, cnv)
	}

	ac.popup.ShowAtPosition(ac.popupPos())
	ac.popup.Resize(ac.popupMaxSize())

	ac.list.Select(0)
	cnv.Focus(ac.list)
}

func (ac *AutoComplete) ListHide() {
	if ac.popup != nil {
		ac.list.UnselectAll()
		ac.popup.Hide()
	}
}

func (ac *AutoComplete) ListVisible() bool {
	return ac.popup != nil && ac.popup.Visible()
}

func (ac *AutoComplete) SetText(s string) {
	ac.pause = true
	ac.Entry.SetText(s)
	ac.pause = false
}

func (ac *AutoComplete) Move(pos fyne.Position) {
	ac.Entry.Move(pos)
	if ac.popup != nil && ac.popup.Visible() {
		ac.popup.Move(ac.popupPos())
		ac.popup.Resize(ac.popupMaxSize())
	}
}

func (ac *AutoComplete) setTextFromList(s string) {
	ac.popup.Hide()
	ac.pause = true
	if ac.OnCompleted != nil {
		s = ac.OnCompleted(s)
	}
	ac.Entry.Text = s
	ac.Entry.CursorColumn = len([]rune(s))
	ac.Entry.Refresh()
	ac.pause = false
	if ac.SubmitOnCompleted && ac.OnSubmitted != nil {
		ac.OnSubmitted(ac.Entry.Text)
	}
}

func (ac *AutoComplete) popupPos() fyne.Position {
	pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(ac)
	return pos.Add(fyne.NewPos(0, ac.Size().Height+theme.Padding()))
}

func (ac *AutoComplete) popupMaxSize() fyne.Size {
	cnv := fyne.CurrentApp().Driver().CanvasForObject(ac)
	if cnv == nil {
		return fyne.Size{}
	}

	pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(ac)

	// define size boundaries
	minWidth := ac.Size().Width
	maxWidth := cnv.Size().Width - pos.X - theme.Padding()
	maxHeight := cnv.Size().Height - pos.Y - ac.MinSize().Height - 2*theme.Padding()

	// iterating items until the end or we reach maxHeight
	var width, height float32
	for i := 0; i < len(ac.Options); i++ {
		item := ac.list.CreateItem()
		ac.list.UpdateItem(i, item)
		sz := item.MinSize()
		if sz.Width > width {
			width = sz.Width
		}
		height += sz.Height + theme.Padding()
		if height > maxHeight {
			height = maxHeight
			break
		}
	}
	height += theme.Padding() // popup padding

	width += 2 * theme.Padding() // let some padding on the triling end of the longest item
	if width < minWidth {
		width = minWidth
	}
	if width > maxWidth {
		width = maxWidth
	}

	return fyne.NewSize(width, height)
}

// ------------------------------------------------------------------------------------------------

type autoCompleteList struct {
	widget.List
	parent *AutoComplete
}

func newAutoCompleteList(parent *AutoComplete) *autoCompleteList {
	list := &autoCompleteList{parent: parent}
	list.ExtendBaseWidget(list)
	list.List.Length = func() int { return len(parent.Options) }
	list.List.CreateItem = func() fyne.CanvasObject {
		var item *autoCompleteListItem
		if parent.CustomCreate != nil {
			item = newAutoCompleteListItem(parent, parent.CustomCreate())
		} else {
			item = newAutoCompleteListItem(parent, widget.NewLabel(""))
		}
		return item
	}
	list.List.UpdateItem = func(id widget.ListItemID, co fyne.CanvasObject) {
		if parent.CustomUpdate != nil {
			parent.CustomUpdate(id, co.(*autoCompleteListItem).co)
		} else {
			co.(*autoCompleteListItem).co.(*widget.Label).Text = parent.Options[id]
		}
		co.(*autoCompleteListItem).id = id
		parent.list.SetItemHeight(id, co.MinSize().Height)
		co.Refresh()
	}
	list.List.OnSelected = func(id widget.ListItemID) {
		parent.selected = id
	}
	list.List.OnUnselected = func(_ widget.ListItemID) {
		parent.selected = -1
	}
	return list
}

func (list *autoCompleteList) AcceptsTab() bool {
	return true
}

func (list *autoCompleteList) FocusGained() {}
func (list *autoCompleteList) FocusLost()   {}

func (list *autoCompleteList) TypedRune(r rune) {
	list.parent.TypedRune(r)
}
func (list *autoCompleteList) TypedKey(k *fyne.KeyEvent) {
	switch k.Name {
	case fyne.KeyDown:
		if list.parent.selected < len(list.parent.Options)-1 {
			list.parent.list.Select(list.parent.selected + 1)
		} else {
			list.parent.list.Select(0)
		}
	case fyne.KeyUp:
		if list.parent.selected > 0 {
			list.parent.list.Select(list.parent.selected - 1)
		} else {
			list.parent.list.Select(len(list.parent.Options) - 1)
		}
	case fyne.KeyReturn, fyne.KeyEnter:
		if list.parent.selected >= 0 {
			list.parent.setTextFromList(list.parent.Options[list.parent.selected])
		} else {
			list.parent.ListHide()
			list.parent.Entry.TypedKey(k)
		}
	case fyne.KeyTab, fyne.KeyEscape:
		list.parent.ListHide()
	default:
		list.parent.TypedKey(k)
	}
}

func (list *autoCompleteList) TypedShortcut(s fyne.Shortcut) { list.parent.TypedShortcut(s) }

// ---

type autoCompleteListItem struct {
	widget.BaseWidget
	parent *AutoComplete
	co     fyne.CanvasObject
	id     widget.ListItemID
}

func newAutoCompleteListItem(parent *AutoComplete, co fyne.CanvasObject) *autoCompleteListItem {
	item := &autoCompleteListItem{parent: parent, id: -1, co: co}
	item.ExtendBaseWidget(item)
	return item
}

func (item *autoCompleteListItem) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(item.co)
}

func (item *autoCompleteListItem) Tapped(_ *fyne.PointEvent) {
	item.parent.setTextFromList(item.parent.Options[item.id])
}

func (item *autoCompleteListItem) MouseIn(_ *desktop.MouseEvent)    { item.parent.list.Select(item.id) }
func (item *autoCompleteListItem) MouseMoved(_ *desktop.MouseEvent) {}
func (item *autoCompleteListItem) MouseOut()                        {}
