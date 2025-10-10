package main

import (
	"bytes"
	"fmt"
	"image"
	"os"

	f "github.com/u2takey/ffmpeg-go"
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
		errorDialog(err.Error())
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
		errorDialog(err.Error())
		panic(err)
	}

	img, _, err := image.Decode(buf)
	if err != nil {
		errorDialog(err.Error())
		panic(err)
	}

	return img
}
