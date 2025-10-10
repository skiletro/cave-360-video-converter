package main

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"

	d "github.com/sqweek/dialog"
)

func stringToInt32(input string) int32 {
	if input == "" {
		return int32(0)
	}

	if output, err := strconv.Atoi(input); err == nil {
		return int32(output)
	} else {
		errorDialog("Not a number. Defaulting to zero.")
		return int32(0)
	}
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

func errorDialog(contents string) {
	d.Message("%s", contents).Title("Cave Video Converter: Error").Error()
}

func infoDialog(contents string) {
	d.Message("%s", contents).Title("Cave Video Converter: Info").Info()
}

func questionDialog(contents string) bool {
	return d.Message("%s", contents).Title("Cave Video Converter: Are you sure?").YesNo()
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

func programIsInstalledInPath(program string) bool {
	_, err := exec.LookPath(program)
	if err == nil {
		return true
	} else {
		return false
	}
}

func checkIfFfmpegIsPresent() {
	if programIsInstalledInPath("ffmpeg") {
		return
	}

	if runtime.GOOS == "windows" && questionDialog("Ffmpeg is not installed. Would you like to install it?") {
		if proc, err := runCommand("winget", "install", "--id=Gyan.FFmpeg", "-e"); err == nil {
			proc.Wait()
			infoDialog("Ffmpeg should now be installed! Please restart the program.")
			os.Exit(1)
		} else {
			errorDialog("Could not install.\n\n" + err.Error())
			os.Exit(1)
		}
	}

	// if we haven't returned yet, we don't have ffmpeg so we should close.
	errorDialog("Ffmpeg is not present in the PATH.\nThis program will not function without it.")
	os.Exit(1)
}
