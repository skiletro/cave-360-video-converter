package main

import (
	"fmt"
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	d "github.com/sqweek/dialog"
	f "github.com/u2takey/ffmpeg-go"
)

var (
	VERSION        string      = "dev"
	LAST_MODIFIED  string      = "unknown"
	img            image.Image = image.NewRGBA(image.Rect(0, 0, 1, 1))
	activityStatus *widget.Activity
)

func main() {
	checkIfFfmpegIsPresent()

	a := app.New()
	w := a.NewWindow(fmt.Sprintf("Cave Video Converter - %s - %s", VERSION, LAST_MODIFIED))

	// Construct UI

	// Construct Preview
	outputWallPreviewImage := canvas.NewImageFromImage(img)
	outputWallPreviewImage.SetMinSize(fyne.NewSize(1280, 240))
	activityStatus = widget.NewActivity()

	previewContainer := container.NewStack(outputWallPreviewImage, activityStatus)

	// Construct Angle
	angleLabel := widget.NewLabel("Angle")
	angleBox := widget.NewEntry()
	angleBox.PlaceHolder = "0"

	angleContainer := container.NewHBox(angleLabel, angleBox)

	// Construct Form Section
	inputLabel := widget.NewLabel("Input")
	inputBox := widget.NewEntry()
	inputSelect := widget.NewButton("Browse", func() {
		if out, err := d.File().Filter("Video Files", "mp4").Title("Load Video").Load(); err == nil {
			inputBox.Text = out
			inputBox.Refresh()
			go previewRoutine(inputBox.Text, stringToInt32(angleBox.Text), outputWallPreviewImage)
		}
	})

	outputLabel := widget.NewLabel("Output")
	outputBox := widget.NewEntry()
	outputSelect := widget.NewButton("Browse", func() {
		if out, err := d.File().Filter("Video Files", "mp4").Title("Output Video Location").Save(); err == nil {
			outputBox.Text = out
			outputBox.Refresh()
		}
	})
	formContainer := container.NewGridWithColumns(3, inputLabel, inputBox, inputSelect, outputLabel, outputBox, outputSelect)

	// Construct Buttons Section
	convertButton := widget.NewButton("Convert", func() {
		if inputBox.Text == "" {
			errorDialog("Input location not specified.")
			return
		}

		if outputBox.Text == "" {
			errorDialog("Output location not specified.")
			return
		}

		go convertRoutine(inputBox.Text, stringToInt32(angleBox.Text), outputBox.Text)
	})
	previewButton := widget.NewButton("Preview", func() {
		if inputBox.Text == "" {
			errorDialog("Input location not specified.")
			return
		}

		go previewRoutine(inputBox.Text, stringToInt32(angleBox.Text), outputWallPreviewImage)
	})

	rotateLeftButton := widget.NewButton("<", func() {
		angleBox.Text = fmt.Sprintf("%d", stringToInt32(angleBox.Text)-int32(10))
		angleBox.Refresh()

		if inputBox.Text != "" {
			go previewRoutine(inputBox.Text, stringToInt32(angleBox.Text), outputWallPreviewImage)
		}
	})

	rotateRightButton := widget.NewButton(">", func() {
		angleBox.Text = fmt.Sprintf("%d", stringToInt32(angleBox.Text)+int32(10))
		angleBox.Refresh()

		if inputBox.Text != "" {
			go previewRoutine(inputBox.Text, stringToInt32(angleBox.Text), outputWallPreviewImage)
		}
	})

	actionButtonsContainer := container.NewCenter(container.NewHBox(rotateLeftButton, convertButton, previewButton, rotateRightButton))

	// Construct Final Layout
	w.SetContent(container.NewVBox(formContainer, angleContainer, actionButtonsContainer, previewContainer))

	w.SetFixedSize(true)
	w.ShowAndRun()
}

// Routines
func previewRoutine(inputPath string, angleInt int32, outputImage *canvas.Image) {
	activityStatus.Show()
	activityStatus.Start()

	inputStream := f.Input(inputPath)

	outputImage.Image = convertWallFrame(inputStream, angleInt)

	outputImage.Refresh()

	activityStatus.Stop()
	activityStatus.Hide()
}

func convertRoutine(inputPath string, angleInt int32, outputPath string) {
	activityStatus.Show()
	activityStatus.Start()

	inputStream := f.Input(inputPath)

	infoDialog("Ready to convert the walls. Continue when ready.")
	convertWalls(inputStream, outputPath+".walls.mp4", angleInt)
	infoDialog("Walls converted sucessfully.\nSaved to " + outputPath + ".walls.mp4")

	infoDialog("Ready to convert the floor. Continue when ready.")
	convertFloor(inputStream, outputPath+".floor.mp4", angleInt)
	infoDialog("Floor converted sucessfully.\nSaved to " + outputPath + ".floor.mp4")

	infoDialog("Conversion complete!")

	activityStatus.Stop()
	activityStatus.Hide()
}
