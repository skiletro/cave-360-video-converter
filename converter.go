package main

import (
	"bytes"
	"fmt"
	"image"
	"os"

	g "github.com/AllenDang/giu"
	"github.com/sqweek/dialog"
	f "github.com/u2takey/ffmpeg-go"
)

var (
	inputLocation   string
	outputLocation  string
	selectedAngle   int32
	previewImageBuf image.Image
	wnd             *g.MasterWindow
)

func filterWallsVisualStream(input *f.Stream, angle int32) *f.Stream {
	left := clampAngle(angle - 90)
	mid := clampAngle(angle)
	right := clampAngle(angle + 90)

	// Convert each wall seperately
	v0 := input.Video().
		Filter("v360", f.Args{"equirect:rectilinear"}, f.KwArgs{"v_fov": 60, "yaw": left}).
		Filter("scale", f.Args{"1920:1080"}).Filter("setsar", f.Args{"1:1"})
	v1 := input.Video().
		Filter("v360", f.Args{"equirect:rectilinear"}, f.KwArgs{"v_fov": 60, "yaw": mid}).
		Filter("scale", f.Args{"1920:1080"}).Filter("setsar", f.Args{"1:1"})
	v2 := input.Video().
		Filter("v360", f.Args{"equirect:rectilinear"}, f.KwArgs{"v_fov": 60, "yaw": right}).
		Filter("scale", f.Args{"1920:1080"}).Filter("setsar", f.Args{"1:1"})

	// Stitch these together
	return f.Filter([]*f.Stream{v0, v1, v2}, "hstack", f.Args{"inputs=3"})
}

func filterFloorVisualStream(input *f.Stream, angle int32) *f.Stream {
	return input.Video().
		Filter("v360", f.Args{"equirect:rectilinear"}, f.KwArgs{"v_fov": 60, "yaw": clampAngle(angle), "pitch": -90}).
		Filter("scale", f.Args{"1920:1920"}).Filter("setsar", f.Args{"1:1"})
}

func convertWalls(inputStream *f.Stream, output string, angle int32) {
	inputVideo := filterWallsVisualStream(inputStream, angle)
	inputAudio := inputStream.Audio()

	out := f.Output([]*f.Stream{inputVideo, inputAudio}, output,
		f.KwArgs{
			"c:a": "copy", // copy over audio unchanged
			"c:v": "h264", // convert video to h264: most widely compatible with intuiface
		})

	cmd := out.OverWriteOutput().Compile()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	} else {
		return
	}
}

func convertFloor(inputStream *f.Stream, output string, angle int32) {
	inputVideo := filterFloorVisualStream(inputStream, angle)

	out := f.Output([]*f.Stream{inputVideo}, output,
		f.KwArgs{
			"c:a": "copy",
			"c:v": "h264",
		})

	cmd := out.OverWriteOutput().Compile()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	} else {
		return
	}
}

func convertWallFrame(inputStream *f.Stream, angle int32) {
	inputVideo := filterWallsVisualStream(inputStream, angle)

	buf := bytes.NewBuffer(nil)
	err := inputVideo.
		Filter("select", f.Args{fmt.Sprintf("gte(n,%d)", 1)}).
		// Filter("scale", f.Args{"1280:240"}).Filter("setsar", f.Args{"1:1"}).
		Output("pipe:", f.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
		WithOutput(buf, os.Stdout).
		Run()
	if err != nil {
		panic(err)
	}

	img, _, err := image.Decode(buf)
	if err != nil {
		panic(err)
	}

	previewImageBuf = img
}

// TODO: convertFloorFrame

func clampAngle(angle int32) int32 {
	// clamp rotation
	clamped := angle % 360

	// convert from range [0, 360] to [-180, 180] with 0 staying in the middle for both
	if clamped < 180 {
		return clamped
	} else {
		return clamped - 360
	}
}

func errBox(text string) {
	dialog.Message("%s", text).Title("Error").Error()
}

func loop() {
	width, height := wnd.GetSize()

	g.SingleWindow().Layout(
		g.Row(
			g.Label("Input"),
			g.InputText(&inputLocation),
			g.Button("Browse").OnClick(func() {
				inputLocation, _ = dialog.File().Filter("Video", "mp4").Title("Load video").Load()
				input := f.Input(inputLocation)
				convertWallFrame(input, selectedAngle)
			}),
		),
		g.Row(
			g.Label("Output"),
			g.InputText(&outputLocation),
			g.Button("Browse").OnClick(func() {
				outputLocation, _ = dialog.File().Filter("Video", "mp4").Title("Output video location").Save()
			}),
		),
		g.Row(
			g.InputInt(&selectedAngle),
		),
		g.Row(
			g.Button("Convert Walls and Floor").OnClick(func() {
				if inputLocation == "" {
					errBox("Input location not selected.")
					return
				}

				if outputLocation == "" {
					errBox("Output location not selected.")
					return
				}
				input := f.Input(inputLocation)

				errBox("Converting! The program might hang until the conversion is complete.")

				convertFloor(input, outputLocation+".floor.mp4", selectedAngle)
				convertWalls(input, outputLocation, selectedAngle)

				errBox("Done!")
			}),
			g.Button("Refresh Preview").OnClick(func() {
				if inputLocation == "" {
					errBox("Input location not selected.")
					return
				}

				input := f.Input(inputLocation)

				convertWallFrame(input, selectedAngle)
			}),
		),
		g.Row(
			g.ImageWithRgba(previewImageBuf).Size(float32(width), float32(height)),
		),
	)
}

func main() {
	previewImageBuf = image.NewRGBA(image.Rect(0, 0, 100, 100))

	wnd = g.NewMasterWindow("Cave 360 Video Converter", 650, 300, 0)
	wnd.Run(loop)
}
