package main

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Task struct {
	Id          uint
	Title       string
	Description string
}

func main() {
	app := app.New()
	app.Settings().SetTheme(theme.LightTheme())
	window := app.NewWindow("ToDo App")
	window.Resize(fyne.NewSize(500, 500))
	window.CenterOnScreen()

	var tasks []Task
	var createContent *fyne.Container
	var tasksContent *fyne.Container
	var tasksList *widget.List

	DB, _ := gorm.Open(sqlite.Open("todo.db"), &gorm.Config{})
	DB.AutoMigrate(&Task{})
	DB.Find(&tasks)

	noTasksLabel := canvas.NewText("No tasks", color.Black)
	noTasksLabel.TextSize = 25
	noTasksLabel.TextStyle = fyne.TextStyle{Monospace: true}
	noTasksLabel.Alignment = fyne.TextAlignCenter

	if len(tasks) != 0 {
		noTasksLabel.Hide()
	}

	newTaskIcon, _ := fyne.LoadResourceFromPath("./icons/plus.png")
	backTaskIcon, _ := fyne.LoadResourceFromPath("./icons/back.png")
	delete, _ := fyne.LoadResourceFromPath("./icons/delete.png")
	edit, _ := fyne.LoadResourceFromPath("./icons/edit.png")
	save, _ := fyne.LoadResourceFromPath("./icons/save.png")

	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Task title...")

	titleEntry.OnChanged = func(text string) {
		for strings.Contains(text, "  ") { // Check for multiple spaces
			text = strings.ReplaceAll(text, "  ", " ") // Replace multiple spaces with a single space
		}
		titleEntry.Text = text // Set the updated text in the TextEntry widget
		titleEntry.Refresh()
	}

	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetPlaceHolder("Task description")

	descriptionEntry.OnChanged = func(text string) {
		for strings.Contains(text, "  ") {
			text = strings.ReplaceAll(text, "  ", " ")
		}
		descriptionEntry.Text = text
		descriptionEntry.Refresh()
	}

	tasksBar := container.NewHBox(
		canvas.NewText("Your tasks", color.Black),
		layout.NewSpacer(),
		widget.NewButtonWithIcon("", newTaskIcon, func() {
			window.SetContent(createContent)
		}),
	)

	tasksList = widget.NewList(
		func() int {
			return len(tasks)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("default")
		},
		func(lii widget.ListItemID, co fyne.CanvasObject) {
			co.(*widget.Label).SetText(tasks[lii].Title)
		},
	)

	tasksList.OnSelected = func(id widget.ListItemID) {

		detailsBar := container.NewHBox(
			canvas.NewText(
				fmt.Sprintf(
					"Details about \"%s\"",
					tasks[id].Title,
				),
				color.Black),
			layout.NewSpacer(),
			widget.NewButtonWithIcon("", backTaskIcon, func() {
				window.SetContent(tasksContent)
				tasksList.Unselect(id)
			}),
		)

		taskTitle := widget.NewLabel(tasks[id].Title)
		taskTitle.TextStyle = fyne.TextStyle{Bold: true}

		taskDescription := widget.NewLabel(tasks[id].Description)
		taskDescription.TextStyle = fyne.TextStyle{Italic: true}
		taskDescription.Wrapping = fyne.TextWrapBreak

		buttonsBox := container.NewHBox(

			//DELETE
			widget.NewButtonWithIcon(
				"",
				delete,
				func() {
					dialog.ShowConfirm(
						"Deleting task",

						fmt.Sprintf("Are you sure about deleting task \"%s\"?", tasks[id].Title),
						func(b bool) {
							if b {
								DB.Delete(&Task{}, "Id", tasks[id].Id)
								DB.Find(&tasks)

								if len(tasks) == 0 {
									noTasksLabel.Show()
								} else {
									noTasksLabel.Hide()
								}
							}

							tasksList.UnselectAll()
							window.SetContent(tasksContent)
						},
						window,
					)
				},
			),
			//EDIT
			widget.NewButtonWithIcon(
				"",
				edit,
				func() {
					editBar := container.NewHBox(
						canvas.NewText(
							fmt.Sprintf(
								"Editing \"%s\"",
								tasks[id].Title,
							),
							color.Black),
						layout.NewSpacer(),
						widget.NewButtonWithIcon("", backTaskIcon, func() {
							window.SetContent(tasksContent)
							tasksList.Unselect(id)
						}),
					)
					editTitle := widget.NewEntry()
					editTitle.SetText(tasks[id].Title)
					editDescription := widget.NewMultiLineEntry()
					editDescription.SetText(tasks[id].Description)

					editTitle.OnChanged = func(text string) {
						for strings.Contains(text, "  ") {
							text = strings.ReplaceAll(text, "  ", " ")
						}
						editTitle.Text = text
						editTitle.Refresh()
					}

					editDescription.OnChanged = func(text string) {
						for strings.Contains(text, "  ") {
							text = strings.ReplaceAll(text, "  ", " ")
						}
						editDescription.Text = text
						editDescription.Refresh()
					}

					editButton := widget.NewButtonWithIcon(
						"Save task",
						save,

						//EDIT TASK FUNCTION
						func() {
							DB.Find(
								&Task{},
								"Id",
								tasks[id].Id,
							).Updates(
								Task{
									Title:       editTitle.Text,
									Description: editDescription.Text,
								},
							)
							DB.Find(&tasks)
							window.SetContent(tasksContent)
							tasksList.UnselectAll()
						},
					)
					editContent := container.NewVBox(
						editBar,
						canvas.NewLine(color.Black),
						editTitle,
						editDescription,
						editButton,
					)
					window.SetContent(editContent)
				},
			),
		)

		detailsVBox := container.NewVBox(
			detailsBar,
			canvas.NewLine(color.Black),
			taskTitle,
			taskDescription,
			buttonsBox,
		)
		window.SetContent(detailsVBox)
	}

	taskScroll := container.NewScroll(tasksList)
	taskScroll.SetMinSize(fyne.NewSize(500, 500))

	tasksContent = container.NewVBox(
		tasksBar,
		canvas.NewLine(color.Black),
		noTasksLabel,
		taskScroll,
	)

	saveTaskButton := widget.NewButtonWithIcon("Save task", newTaskIcon, func() {

		if (strings.TrimSpace(titleEntry.Text)) != "" {

			task := Task{Title: titleEntry.Text, Description: descriptionEntry.Text}
			DB.Create(&task)
			DB.Find(&tasks)
			titleEntry.Text = ""
			titleEntry.Refresh()
			descriptionEntry.Text = ""
			descriptionEntry.Refresh()

			window.SetContent(tasksContent)

			tasksList.UnselectAll()

			if len(tasks) == 0 {
				noTasksLabel.Show()
			} else {
				noTasksLabel.Hide()
			}
		}
	})

	createBar := container.NewHBox(
		canvas.NewText("Create new task", color.Black),
		layout.NewSpacer(),
		widget.NewButtonWithIcon("", backTaskIcon, func() {
			titleEntry.Text = ""
			titleEntry.Refresh()
			descriptionEntry.Text = ""
			descriptionEntry.Refresh()
			window.SetContent(tasksContent)
			tasksList.UnselectAll()

		}),
	)

	createContent = container.NewVBox(
		createBar,
		canvas.NewLine(color.Black),
		container.NewVBox(
			titleEntry,
			descriptionEntry,
			saveTaskButton,
		),
	)

	window.SetContent(tasksContent)
	window.Show()
	app.Run()

}
