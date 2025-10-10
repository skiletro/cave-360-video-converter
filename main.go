package main

import (
	"image"

	g "github.com/AllenDang/giu"
	"github.com/sqweek/dialog"
	f "github.com/u2takey/ffmpeg-go"
)

var (
	inputLocation     string
	outputLocation    string
	selectedAngle     int32
	previewWallsImage image.Image
	previewFloorimage image.Image
	wnd               *g.MasterWindow
)

func loop() {
	width, _ := wnd.GetSize()

	g.SingleWindow().Layout(
		g.Align(g.AlignCenter).To(
			g.Row(
				g.Label("Input"),
				g.InputText(&inputLocation),
				g.Button("Browse").OnClick(func() {
					var err error
					inputLocation, err = dialog.File().Filter("Video files", "mp4").Title("Load video").Load()

					if err == nil {
						input := f.Input(inputLocation)
						convertWallFrame(input, selectedAngle)
					}
				}),
			),
			g.Row(
				g.Label("Output"),
				g.InputText(&outputLocation),
				g.Button("Browse").OnClick(func() {
					outputLocation, _ = dialog.File().Filter("Video files", "mp4").Title("Output video location").Save()
				}),
			),
			g.Row(
				g.Label("Rotational Angle"),
				g.InputInt(&selectedAngle),
			),
			g.Row(
				g.Button("Convert Walls and Floor").OnClick(func() {
					if inputLocation == "" {
						errorDialog("Input location not selected.")
						return
					}

					input := f.Input(inputLocation)

					if outputLocation == "" {
						errorDialog("Output location not selected.")
						return
					}

					infoDialog("Converting!\nThe program might hang until the conversion is complete.")

					convertWalls(input, outputLocation+".walls.mp4", selectedAngle)
					convertFloor(input, outputLocation+".floor.mp4", selectedAngle)

					infoDialog("Done!")
				}),
				g.Button("Refresh Preview").OnClick(func() {
					if inputLocation == "" {
						errorDialog("Input location not selected.")
						return
					}

					input := f.Input(inputLocation)

					convertWallFrame(input, selectedAngle)
				}),
			),
			g.Row(
				g.ImageWithRgba(previewWallsImage).Size(float32(width), float32(width)*(9.0/48.0)),
			),
		),
	)
}

func main() {
	checkIfFfmpegIsPresent()

	previewWallsImage = image.NewRGBA(image.Rect(0, 0, 100, 100))

	wnd = g.NewMasterWindow("Cave 360 Video Converter", 650, 300, 0)
	wnd.Run(loop)
}
