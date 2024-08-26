package media

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

type Format struct {
	Filename       string
	Streams        int
	Programs       int
	Extensions     []string
	FormatLongName string
	Duration       time.Duration
	Size           uint
	BitRate        uint
	ProbeScore     int
	Tags           Tags
}

type basicFormat struct {
	Filename       string
	NbStreams      int    `json:"nb_streams"`
	NbPrograms     int    `json:"nb_programs"`
	FormatName     string `json:"format_name"`
	FormatLongName string `json:"format_long_name"`
	Duration       string `json:"duration"`
	Size           string `json:"size"`
	BitRate        string `json:"bit_rate"`
	ProbeScore     int    `json:"probe_score"`
	Tags           Tags   `json:"tags"`
}

func (f *Format) UnmarshalJSON(bytes []byte) error {
	fmt := new(basicFormat)
	if err := json.Unmarshal(bytes, fmt); err != nil {
		return err
	}

	f.Filename = fmt.Filename
	f.Extensions = strings.Split(fmt.FormatName, ",")
	f.Streams = fmt.NbStreams
	f.Programs = fmt.NbPrograms
	f.Tags = fmt.Tags
	f.ProbeScore = fmt.ProbeScore

	if bitRate, err := strconv.Atoi(fmt.BitRate); err == nil {
		f.BitRate = uint(bitRate)
	}
	if size, err := strconv.Atoi(fmt.Size); err == nil {
		f.Size = uint(size)
	}
	if dur, err := time.ParseDuration(fmt.Duration + "s"); err == nil {
		f.Duration = dur
	}

	return nil
}

func (f *Format) Seconds() float64 {
	return float64(f.Duration.Nanoseconds()) / float64(time.Second)
}

type Tags struct {
	Encoder string `json:"ENCODER"`
}

type (
	Orientation string
	CodecType   string
)

const (
	CodecTypeVideo       CodecType   = "video"
	CodecTypeAudio       CodecType   = "audio"
	OrientationLandscape Orientation = "landscape"
	OrientationPortrait  Orientation = "portrait"
)
