package utils

import (
	"bytes"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/graux/goffmpeg/models"
)

const (
	execNameFFMpeg  = "ffmpeg"
	execNameFFProbe = "ffprobe"
	osWindows       = "windows"
	commandWhich    = "which"
	commandWhere    = "where"
)

func DurToSec(dur string) (sec float64) {
	durAry := strings.Split(dur, ":")
	var secs float64
	if len(durAry) != 3 {
		return
	}
	hr, _ := strconv.ParseFloat(durAry[0], 64)
	secs = hr * (60 * 60)
	min, _ := strconv.ParseFloat(durAry[1], 64)
	secs += min * (60)
	second, _ := strconv.ParseFloat(durAry[2], 64)
	secs += second
	return secs
}

func GetFFmpegExec() []string {
	return []string{getLocateBinCommand(), execNameFFMpeg}
}

func GetFFprobeExec() []string {
	return []string{getLocateBinCommand(), execNameFFProbe}
}

func isWindows() bool {
	return runtime.GOOS == osWindows
}

func getLocateBinCommand() string {
	if isWindows() {
		return commandWhere
	}
	return commandWhich
}

func CheckFileType(streams []models.Streams) models.CodecType {
	for i := 0; i < len(streams); i++ {
		st := streams[i]
		if st.CodecType == models.CodecTypeVideo {
			return models.CodecTypeVideo
		}
	}
	return models.CodecTypeAudio
}

func LineSeparator() string {
	if isWindows() {
		return "\r\n"
	}
	return "\n"
}

// TestCmd ...
func TestCmd(command string, args string) (bytes.Buffer, error) {
	var out bytes.Buffer

	cmd := exec.Command(command, args)

	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return out, err
	}

	return out, nil
}
