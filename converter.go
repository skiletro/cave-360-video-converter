package main

import (
	"os"

	g "github.com/AllenDang/giu"
	"github.com/sqweek/dialog"
	f "github.com/u2takey/ffmpeg-go"
)

var (
	inputLocation  string
	outputLocation string
	selectedAngle  int32
)

func filterWallsAudioVisualStreams(input *f.Stream, angle int32) *f.Stream {
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

func convertVideo(inputStream *f.Stream, output string, angle int32) {
	inputVideo := filterWallsAudioVisualStreams(inputStream, angle)
	inputAudio := inputStream.Audio()

	out := f.Output([]*f.Stream{inputVideo, inputAudio}, output,
		f.KwArgs{
			"c:a": "copy", // copy over audio unchanged
			"c:v": "h264", // convert video to h264: most widely compatible with intuiface
			"t":   "1",    // TODO: remove this, just renders out the first second.
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

func convertFrame(inputStream *f.Stream, output string, angle int32) {
	inputVideo := filterWallsAudioVisualStreams(inputStream, angle)

	out := f.Output([]*f.Stream{inputVideo}, output,
		f.KwArgs{
			"c:v":     "mjpeg",
			"vframes": 1,
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

func clampAngle(angle int32) int32 {
	// clamp rotation
	clamped := angle % 360

	// convert from range [0, 360] to [-180, 180] with 0 staying in the middle for both
	var ranged int32
	if clamped < 180 {
		ranged = clamped
	} else {
		ranged = clamped - 360
	}

	// reverse (for whatever reason, ffmpeg likes it like this)
	return ranged
}

func errBox(text string) {
	dialog.Message("%s", text).Title("Error").Error()
}

func loop() {
	g.SingleWindow().Layout(
		g.Row(
			g.Label("Input"),
			g.InputText(&inputLocation),
			g.Button("Browse").OnClick(func() {
				inputLocation, _ = dialog.File().Filter("Video", "mp4").Title("Load video").Load()
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
			g.Button("Convert").OnClick(func() {
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

				convertVideo(input, outputLocation, selectedAngle)

				errBox("Done!")
			}),
		),
	)
}

func main() {
	wnd := g.NewMasterWindow("On Drop Demo", 600, 200, g.MasterWindowFlagsNotResizable)
	wnd.Run(loop)
}
