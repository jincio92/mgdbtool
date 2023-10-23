package main

import (
	"fmt"
	"jincio/mgdbtool/internal/db"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type MainForm struct {
	mainWindow        *fyne.Window
	dbFromEntry       *widget.Entry
	userFromEntry     *widget.Entry
	passwordFromEntry *widget.Entry
	dbToEntry         *widget.Entry
	userToEntry       *widget.Entry
	passwordToEntry   *widget.Entry
	filterEntry       *widget.Entry
	progressItem      *widget.FormItem
	scrollItem        *widget.FormItem
	scanButton        *widget.Button
	formItems         []*widget.FormItem
	checkGroup        *widget.CheckGroup
	OnSubmit          func()
	OnCancel          func()
}

func NewMainForm(window *fyne.Window) *MainForm {
	form := &MainForm{mainWindow: window}
	form.setupUI()
	return form
}

func (form *MainForm) setupUI() {
	form.dbFromEntry = widget.NewEntry()
	form.dbFromEntry.SetPlaceHolder("prod.domain:5432/postgres")
	form.userFromEntry = widget.NewEntry()
	form.userFromEntry.SetPlaceHolder("User DB From")
	form.passwordFromEntry = widget.NewEntry()
	form.passwordFromEntry.SetPlaceHolder("Password")

	form.dbToEntry = widget.NewEntry()
	form.dbToEntry.SetPlaceHolder("localhost:5432/postgres")
	form.userToEntry = widget.NewEntry()
	form.userToEntry.SetPlaceHolder("User DB To")
	form.passwordToEntry = widget.NewEntry()
	form.passwordToEntry.SetPlaceHolder("Password")

	form.filterEntry = widget.NewMultiLineEntry()
	form.filterEntry.SetPlaceHolder("column_name = 'foo'")

	form.checkGroup = widget.NewCheckGroup([]string{}, nil)

	form.progressItem = widget.NewFormItem("", widget.NewProgressBarInfinite())
	form.progressItem.Widget.Hide()
	form.scrollItem = widget.NewFormItem("Table List", container.NewScroll(container.NewVBox(form.checkGroup)))
	form.scrollItem.Widget.Hide()

	form.scanButton = widget.NewButtonWithIcon("Scan", theme.SearchIcon(), form.handleScanButton)
	form.formItems = []*widget.FormItem{
		{Text: "DB From", Widget: form.dbFromEntry, HintText: "The DB From which read the data"},
		{Text: "DB From User", Widget: form.userFromEntry, HintText: "The user of the From DB"},
		{Text: "DB From Password", Widget: form.passwordFromEntry, HintText: "The password of the FromDB"},
		{Text: "Scan table", Widget: form.scanButton, HintText: "test"},
		form.progressItem,
		form.scrollItem,
		{Widget: widget.NewSeparator()},
		{Text: "DB To", Widget: form.dbToEntry, HintText: "The DB to write the data in"},
		{Text: "DB From User", Widget: form.userToEntry, HintText: "The user of the From DB"},
		{Text: "DB From Password", Widget: form.passwordToEntry, HintText: "The password of the FromDB"},
		{Widget: widget.NewSeparator()},
		{Text: "Filter", Widget: form.filterEntry, HintText: "this will be place after the WHERE statement, es column_name = 'foo'"},
	}

	form.setEventHandlers()
}

func (form *MainForm) setEventHandlers() {
	form.OnCancel = func() {

		fmt.Println("Cancelled")
		os.Exit(1)
	}
	form.OnSubmit = func() {
		fmt.Println("Form submitted")
		configFrom := db.DatabaseConfig{URL: form.dbFromEntry.Text, User: form.userFromEntry.Text, Password: form.passwordFromEntry.Text}
		configTo := db.DatabaseConfig{URL: form.dbToEntry.Text, User: form.userToEntry.Text, Password: form.passwordToEntry.Text}
		if !strings.Contains(form.dbToEntry.Text, "localhost") {

			cnf := dialog.NewConfirm("Warning", "The address of the destination DB does not contain 'localhost',\nare you sure you are pointing to the correct DB?", func(confirm bool) {
				if confirm {
					err := db.ConnectToSql(configFrom, configTo, form.filterEntry.Text, form.checkGroup.Selected)
					if err != nil {
						dialog.ShowError(err, *form.mainWindow)
					}
				}
			}, *form.mainWindow)
			cnf.SetDismissText("Cancel")
			cnf.SetConfirmText("Proceed")
			cnf.Show()
		} else {
			err := db.ConnectToSql(configFrom, configTo, form.filterEntry.Text, form.checkGroup.Selected)
			if err != nil {
				dialog.ShowError(err, *form.mainWindow)
			}
		}
	}
}

func (form *MainForm) handleScanButton() {
	form.progressItem.Widget.Show()
	config := db.DatabaseConfig{URL: form.dbFromEntry.Text, User: form.userFromEntry.Text, Password: form.passwordFromEntry.Text}
	list, err := db.ListTables(config)
	if err != nil {
		dialog.ShowError(err, *form.mainWindow)
	}
	form.checkGroup.Options = list
	form.checkGroup.Refresh()
	if scroll, ok := form.scrollItem.Widget.(*container.Scroll); ok {
		scroll.SetMinSize(fyne.NewSize(60, 250))
	}
	form.progressItem.Widget.Hide()
	form.scrollItem.Widget.Show()
}
func (form *MainForm) createForm() *widget.Form {
	return &widget.Form{
		Items:    form.formItems,
		OnSubmit: form.OnSubmit,
		OnCancel: form.OnCancel}
}

func main() {

	a := app.New()
	w := a.NewWindow("DB Import Data Tool")

	mainForm := NewMainForm(&w)
	w.SetContent(mainForm.createForm())
	w.Resize(fyne.NewSize(800, 900))
	w.ShowAndRun()
}
