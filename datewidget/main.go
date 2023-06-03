package main

import (
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Date Entry")

	date := NewDateEntry()
	date.OnSubmitted = func(tm time.Time) {
		dialog.ShowInformation("Date Entry", tm.Format("Entered date: 02/01/2006"), w)
	}

	w.SetContent(container.NewBorder(date, nil, nil, nil, layout.NewSpacer()))

	w.ShowAndRun()
}

// DateEntry is a widget.Entry that only accept a date input.
// Date format is 02/01/2006 (dd/mm/yyyy).
type DateEntry struct {
	widget.Entry

	OnChanged   func(time.Time) // Called when the data changes
	OnSubmitted func(time.Time) // Called when Enter is pressed in the input

	valid bool
}

// NewDateEntry creates a new DateEntry.
func NewDateEntry() *DateEntry {
	d := &DateEntry{}
	d.ExtendBaseWidget(d)
	d.Text = "__/__/____"
	d.Entry.OnSubmitted = func(s string) {
		if d.OnSubmitted != nil {
			d.OnSubmitted(d.readTime())
		}
	}
	d.Entry.OnChanged = func(s string) {
		d.Validate()

		tm := d.readTime()
		valid := !tm.IsZero()

		if valid != d.valid {
			if d.OnChanged != nil {
				d.OnChanged(tm)
			}
			d.valid = valid
		}
	}
	return d
}

func (d *DateEntry) callOnChanged() {
	d.Entry.OnChanged(d.Text)
}

// SetString will set date entry text.
// Only dates in format 02/01/2006 are accepted. Any other caracter in s will be ignored.
func (d *DateEntry) SetString(s string) {
	d.Text = "__/__/____"
	d.CursorColumn = 0
	for _, r := range s {
		d.TypedRune(r)
	}
	d.valid = !d.readTime().IsZero()
	d.Refresh()
}

// GetString returns the currently entered date in string format (02/01/2006).
// If entered date is not valid, it will return empty string.
func (d *DateEntry) GetString() string {
	tm, err := time.ParseInLocation("02/01/2006", d.Text, time.Local)
	if err != nil || tm.IsZero() {
		return ""
	}
	return d.Text
}

// SetTime will set currently displayed date to tm.
// If tm.IsZero(), it will set empty date (__/__/____).
func (d *DateEntry) SetTime(tm time.Time) {
	if tm.IsZero() {
		d.Text = "__/__/____"
		d.CursorColumn = 0
	} else {
		d.Text = tm.Format("02/01/2006")
		d.CursorColumn = 10
	}
	d.valid = !d.readTime().IsZero()
	d.Refresh()
}

// GetTime wil return the currently entered date as time.Time.
// If entered date is not valid, it will return a zero time object.
func (d *DateEntry) GetTime() (tm time.Time) {
	tm, _ = time.ParseInLocation("02/01/2006", d.Text, time.Local)
	return
}

func (d *DateEntry) MinSize() fyne.Size {
	s := d.Entry.MinSize()
	s.Width = fyne.MeasureText("00/00/0000", theme.TextSize(), d.TextStyle).Width + 2*theme.InnerPadding() + 2*theme.InputBorderSize()
	return s
}

func (d *DateEntry) TypedRune(r rune) {
	switch r {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		switch d.CursorColumn {
		// __/__/____ 0, 1, /2, 3, 4, /5, 6, 7, 8, 9 [, 10]
		case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9:
			t := []rune(d.Text)
			if d.CursorColumn == 2 || d.CursorColumn == 5 {
				d.CursorColumn += 1
			}
			t[d.CursorColumn] = r
			d.CursorColumn += 1
			if d.CursorColumn == 2 || d.CursorColumn == 5 {
				d.CursorColumn += 1
			}
			d.Text = string(t)
			d.callOnChanged()
			d.Refresh()
		}
	}
}

func (d *DateEntry) TypedKey(k *fyne.KeyEvent) {
	switch k.Name {
	case fyne.KeyRight:
		d.CursorColumn += 1
		if d.CursorColumn >= 10 {
			d.CursorColumn = 10
		}
		if d.CursorColumn == 2 || d.CursorColumn == 5 {
			d.CursorColumn += 1
		}
	case fyne.KeyLeft:
		d.CursorColumn -= 1
		if d.CursorColumn <= 0 {
			d.CursorColumn = 0
		}
		if d.CursorColumn == 2 || d.CursorColumn == 5 {
			d.CursorColumn -= 1
		}
	case fyne.KeyUp:
		switch d.CursorColumn {
		case 0, 1, 2:
			d.setDay(d.getDay()+1, true)
		case 3, 4, 5:
			d.setMonth(d.getMonth()+1, true)
		case 6, 7, 8, 9, 10:
			d.setYear(d.getYear() + 1)
		}
		d.callOnChanged()
	case fyne.KeyDown:
		switch d.CursorColumn {
		case 0, 1, 2:
			d.setDay(d.getDay()-1, true)
		case 3, 4, 5:
			d.setMonth(d.getMonth()-1, true)
		case 6, 7, 8, 9, 10:
			d.setYear(d.getYear() - 1)
		}
		d.callOnChanged()
	case fyne.KeyBackspace:
		// __/__/____ 0, 1, /2, 3, 4, /5, 6, 7, 8, 9 [, 10]
		t := []rune(d.Text)
		switch d.CursorColumn {
		case 10, 9, 8, 7, 5, 4, 2, 1:
			t[d.CursorColumn-1] = '_'
			d.CursorColumn -= 1
			if d.CursorColumn == 3 || d.CursorColumn == 6 {
				d.CursorColumn -= 1
			}
		case 3, 6:
			d.CursorColumn -= 1
		}
		d.Text = string(t)
		d.callOnChanged()
	case fyne.KeyEnter, fyne.KeyReturn:
		d.Entry.TypedKey(k)
	case fyne.KeyDelete, fyne.KeyEscape:
		d.Text = "__/__/____"
		d.CursorColumn = 0
		d.callOnChanged()
	default:
		return
	}
	d.Refresh()
}

func (d *DateEntry) TypedShortcut(shortcut fyne.Shortcut) {
	if s, ok := shortcut.(*fyne.ShortcutPaste); ok {
		for _, r := range s.Clipboard.Content() {
			d.TypedRune(r)
		}
	} else {
		d.Entry.TypedShortcut(shortcut)
	}
}

// ------------------------------------------------------------------------------------------------

func (d *DateEntry) readTime() time.Time {
	tm, _ := time.ParseInLocation("02/01/2006", d.Text, time.Local)
	return tm
}

func (d *DateEntry) setDay(day int, loop bool) {
	maxDay := 30
	switch d.getMonth() {
	case 4, 6, 9, 11:
		maxDay = 30
	case 1, 3, 5, 7, 8, 10, 12:
		maxDay = 31
	case 2:
		year := d.getYear()
		if year == 0 {
			maxDay = 28
		} else {
			if year%400 == 0 || (year%4 == 0 && year%100 != 0) {
				maxDay = 29
			} else {
				maxDay = 28
			}
		}
	}
	if day > maxDay {
		if loop {
			day = 1
		} else {
			day = maxDay
		}
	}
	if day < 1 {
		if loop {
			day = maxDay
		} else {
			day = 1
		}
	}
	d.Text = fmt.Sprintf("%02d", day) + d.Text[2:]
}
func (d *DateEntry) getDay() int {
	ret, err := strconv.Atoi(d.Text[:2]) // __/__/____
	if err != nil {
		return 0
	}
	return ret
}

func (d *DateEntry) setMonth(month int, loop bool) {
	if month > 12 {
		if loop {
			month = 1
		} else {
			month = 12
		}
	}
	if month < 1 {
		if loop {
			month = 12
		} else {
			month = 1
		}
	}
	sMonth := fmt.Sprintf("%02d", month)
	d.Text = d.Text[:3] + sMonth + d.Text[5:]
}
func (d *DateEntry) getMonth() int {
	ret, err := strconv.Atoi(d.Text[3:5]) // __/__/____
	if err != nil {
		return 0
	}
	return ret
}

func (d *DateEntry) setYear(year int) {
	if year > 9999 {
		year = 9999
	}
	if year < 1 {
		year = 1
	}
	sYear := fmt.Sprintf("%04d", year)
	d.Text = d.Text[:6] + sYear
}
func (d *DateEntry) getYear() int {
	ret, err := strconv.Atoi(d.Text[6:]) // __/__/____
	if err != nil {
		return 0
	}
	return ret
}
