package main

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"os/exec"
	"runtime"

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

func applyFilterToWallStream(input *f.Stream, angle int32) *f.Stream {
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

func applyFilterToFloorStream(input *f.Stream, angle int32) *f.Stream {
	return input.Video().
		Filter("v360", f.Args{"equirect:rectilinear"}, f.KwArgs{"v_fov": 60, "yaw": clampAngle(angle), "pitch": -90}).
		Filter("scale", f.Args{"1920:1920"}).Filter("setsar", f.Args{"1:1"})
}

func convertWalls(inputStream *f.Stream, outputFile string, angle int32) {
	saveStreamsToFile(applyFilterToWallStream(inputStream, angle), inputStream.Audio(), outputFile)
}

func convertFloor(inputStream *f.Stream, outputFile string, angle int32) {
	saveStreamsToFile(applyFilterToFloorStream(inputStream, angle), inputStream.Audio(), outputFile)
}

func saveStreamsToFile(videoStream *f.Stream, audioStream *f.Stream, outputFile string) {
	out := f.Output([]*f.Stream{videoStream, audioStream}, outputFile,
		f.KwArgs{
			"c:a": "copy", // copy over audio unchanged
			"c:v": "h264", // convert video to h264: most widely compat. with intuiface
		})

	cmd := out.OverWriteOutput().Compile()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		dialogBox(err.Error(), true)
		panic(err)
	} else {
		return
	}
}

func convertWallFrame(inputStream *f.Stream, angle int32) {
	inputVideo := applyFilterToWallStream(inputStream, angle)
	previewWallsImage = streamToFirstFrameAsImage(inputVideo)
}

func streamToFirstFrameAsImage(stream *f.Stream) image.Image {
	buf := bytes.NewBuffer(nil)

	err := stream.
		Filter("select", f.Args{fmt.Sprintf("gte(n,%d)", 1)}).
		Output("pipe:", f.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
		WithOutput(buf, os.Stdout).
		Run()
	if err != nil {
		dialogBox(err.Error(), true)
		panic(err)
	}

	img, _, err := image.Decode(buf)
	if err != nil {
		dialogBox(err.Error(), true)
		panic(err)
	}

	return img
}

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

func dialogBox(text string, isError bool) {
	d := dialog.Message("%s", text).Title("Cave Converter")

	if isError {
		d.Error()
	} else {
		d.Info()
	}
}

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
						dialogBox("Input location not selected.", true)
						return
					}

					input := f.Input(inputLocation)

					if outputLocation == "" {
						dialogBox("Output location not selected.", true)
						return
					}

					dialogBox("Converting! The program might hang until the conversion is complete.", false)

					convertWalls(input, outputLocation+".walls.mp4", selectedAngle)
					convertFloor(input, outputLocation+".floor.mp4", selectedAngle)

					dialogBox("Done!", false)
				}),
				g.Button("Refresh Preview").OnClick(func() {
					if inputLocation == "" {
						dialogBox("Input location not selected.", true)
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

func ffmpegIsInstalledInPath() bool {
	_, err := exec.LookPath("ffmpeg")
	if err == nil {
		return true
	} else {
		return false
	}
}

func runCommand(args ...string) (p *os.Process, err error) {
	if args[0], err = exec.LookPath(args[0]); err == nil {
		var procAttr os.ProcAttr
		procAttr.Files = []*os.File{
			os.Stdin,
			os.Stdout, os.Stderr,
		}
		p, err := os.StartProcess(args[0], args, &procAttr)
		if err == nil {
			return p, nil
		}
	}
	return nil, err
}

func checkIfFfmpegIsPresent() {
	if ffmpegIsInstalledInPath() {
		return
	}

	if runtime.GOOS == "windows" && dialog.Message("%s", "Ffmpeg is not installed. Would you like to install it?").Title("Are you sure?").YesNo() {
		if proc, err := runCommand("winget", "install", "--id=Gyan.FFmpeg", "-e"); err == nil {
			proc.Wait()
			dialogBox("Ffmpeg should now be installed! Please restart the program.", false)
			os.Exit(1)
		} else {
			dialogBox("Could not install.\n\n"+err.Error(), true)
			os.Exit(1)
		}
	}

	// if we haven't returned yet, we don't have ffmpeg so we should close.
	dialogBox("Ffmpeg is not present in the PATH.\nThis program will not function without it.", true)
	os.Exit(1)
}

func main() {
	checkIfFfmpegIsPresent()

	previewWallsImage = image.NewRGBA(image.Rect(0, 0, 100, 100))

	wnd = g.NewMasterWindow("Cave 360 Video Converter", 650, 300, 0)
	wnd.Run(loop)
}
