package main

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/brianvoe/gofakeit/v6"
)

/*
This example show how to handle complexe list items in fyne.
*/

func main() {
	a := app.New()
	w := a.NewWindow("Fyne complexe list item example")

	var section UISectionList
	section.init()

	// some controls
	rowid := int64(1)
	btnAddRandom := &widget.Button{
		Text: "Add Random Element",
		OnTapped: func() {
			// just append some random data
			section.data = append(section.data, listData{
				DBData: DBData{
					RowID:     rowid,
					FirstName: gofakeit.FirstName(),
					LastName:  gofakeit.LastName(),
					DOB:       gofakeit.DateRange(time.Date(1920, 1, 1, 0, 0, 0, 0, time.Local), time.Now()),
				},
			})
			// and refresh the list to display it
			section.list.Refresh()
			rowid += 1
		},
	}

	output := widget.NewMultiLineEntry()

	btnGetSelection := &widget.Button{
		Text: "Get Selected Items",
		OnTapped: func() {
			// you do what you want with selection items...
			sel := section.getSelection()

			// ... we will just display them in output Entry
			output.Text = ""
			for i := 0; i < len(sel); i++ {
				// not optimized yes... I know...
				output.Text += fmt.Sprintf("%s %s (%s)\n", sel[i].FirstName, sel[i].LastName, sel[i].DOB.Format("02/01/2006"))
			}
			output.Refresh()
		},
	}

	// layout the widgets in the window
	// remember: when you are stuck and unable to layout as you want, NewBorder is nearly always the answer!
	w.SetContent(container.NewGridWithColumns(2,
		container.NewBorder(
			btnAddRandom,
			container.NewHBox(section.multiSel, layout.NewSpacer(), btnGetSelection),
			nil, nil,
			section.list,
		),
		output,
	))

	w.Resize(fyne.NewSize(600, 400))
	w.CenterOnScreen()
	w.ShowAndRun()
}

// let's define some data structure that will represent the data of a list item
// this could be for example some DataBase object

type DBData struct {
	RowID               int64
	FirstName, LastName string
	DOB                 time.Time
}

// then, in order to not modifiy the DBObject structure, we create an "extended" structure
// that will hold additionnal data usefull to fyne : the selected marker

type listData struct {
	DBData

	selected bool
}

// I like placing all related widgets and their data in a single struct
// it's easier to handle complexe UIs like this
// and you can reuse the same components in many places in your app

type UISectionList struct {
	list     *widget.List      // the actuel list widget
	data     []listData        // list items
	selID    widget.ListItemID // track currently selected item
	multiSel *widget.Check     // multiselection switch
}

func (section *UISectionList) init() {
	// create the list widget, with its callbacks
	section.list = widget.NewList(
		// this tells the list how many items there are in data sclice
		func() int { return len(section.data) },

		// this one is used by the list to create a canvasObject it will use a list item
		func() fyne.CanvasObject { return newListItem(section) },

		// this one is used by the list to populate data in an item canvasObject
		// DO NOT create items here!
		func(id widget.ListItemID, co fyne.CanvasObject) { co.(*listItem).update(id) },
	)

	// just track selection
	section.list.OnSelected = func(id widget.ListItemID) {
		section.selID = id
	}
	section.list.OnUnselected = func(_ widget.ListItemID) {
		section.selID = -1
	}

	// create the multi-selection switch
	section.multiSel = widget.NewCheck("Multi-selection", func(b bool) {
		// if setting multisel, reset all items selection state
		if b {
			for i := 0; i < len(section.data); i++ {
				section.data[i].selected = false
			}
		}

		// unselect list items
		section.list.UnselectAll()

		// refresh the list, so all items will be update according to multisel status
		section.list.Refresh()
	})
}

// this function will return a slice of all selected items,
// either in single selection mode or multiselection
func (section *UISectionList) getSelection() (ret []*DBData) {
	if section.multiSel.Checked {
		for i := 0; i < len(section.data); i++ {
			if section.data[i].selected {
				ret = append(ret, &section.data[i].DBData)
			}
		}
	} else {
		if section.selID >= 0 {
			ret = append(ret, &section.data[section.selID].DBData)
		}
	}
	return
}

// this is the custom list item, where all the magic happens!

type listItem struct {
	widget.BaseWidget

	parent *UISectionList    // store a ref to the parent tab section
	id     widget.ListItemID // each item should know where it belongs in the backing data

	chk    *widget.Check  // check selection, initially hidden
	icon   *widget.Icon   // just some icon
	lbl1   *widget.Label  // will hold the name
	lbl2   *widget.Label  // will hold DOB
	delBtn *widget.Button // a button to delete the item from data slice
}

func newListItem(parent *UISectionList) (item *listItem) {
	item = &listItem{parent: parent, id: -1}

	// dont forget to extend base widget
	item.ExtendBaseWidget(item)

	// the selection checkbox widget
	item.chk = widget.NewCheck("", func(b bool) {
		// update self selection state
		item.parent.data[item.id].selected = b
	})

	// some icon, it will be hidden and replaced by the selection check when multiSel.Checked = true
	item.icon = widget.NewIcon(theme.AccountIcon())

	// data labels
	item.lbl1 = widget.NewLabel("")
	item.lbl2 = widget.NewLabelWithStyle("", fyne.TextAlignTrailing, fyne.TextStyle{Italic: true})

	// delete button
	item.delBtn = &widget.Button{
		Icon:       theme.DeleteIcon(),
		Importance: widget.LowImportance,
		OnTapped: func() {
			// remove this data item
			item.parent.data = append(item.parent.data[:item.id], item.parent.data[item.id+1:]...)

			// reset list selection if the deleted item was selected
			if item.id == item.parent.selID {
				item.parent.list.Unselect(item.id)
			}

			// and refresh the list
			item.parent.list.Refresh()
		},
	}

	//
	return
}

func (item *listItem) update(id widget.ListItemID) {
	item.id = id

	// update the icon and the selection check
	if item.parent.multiSel.Checked {
		item.chk.Show()
		item.icon.Hide()
		item.chk.SetChecked(item.parent.data[item.id].selected) // update check status
	} else {
		item.chk.Hide()
		item.icon.Show()
	}

	// update labels
	item.lbl1.Text = fmt.Sprintf("%s %s", item.parent.data[item.id].FirstName, item.parent.data[item.id].LastName)
	item.lbl2.Text = item.parent.data[item.id].DOB.Format("02/01/2006")

	// refresh everything at once (this is why we don't use lbl.SetText)
	item.Refresh()
}

// (bonus) let's say we don't want items to be selectable when multiselect is ON
func (item *listItem) Tapped(_ *fyne.PointEvent) {
	// override Tapped handler of items
	if !item.parent.multiSel.Checked {
		// and call list.Select only if not in multiselect mode
		item.parent.list.Select(item.id)
	}
}

func (item *listItem) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewHBox(
		container.NewMax(item.chk, item.icon), // or container.NewStack
		item.lbl1, layout.NewSpacer(), item.lbl2,
		widget.NewSeparator(), item.delBtn,
	))
}
