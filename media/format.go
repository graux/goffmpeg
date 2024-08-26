package media

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

type Format struct {
	Filename       string
	Streams        int    `json:"nb_streams"`
	Programs       int    `json:"nb_programs"`
	FormatName     string `json:"format_name"`
	Extensions     []string
	FormatLongName string `json:"format_long_name"`
	DurationStr    string `json:"duration"`
	Duration       time.Duration
	Size           uint
	BitRate        uint
	SizeStr        string `json:"size"`
	BitRateStr     string `json:"bit_rate"`
	ProbeScore     int    `json:"probe_score"`
	Tags           Tags   `json:"tags"`
}

func (f *Format) UnmarshalJSON(bytes []byte) error {
	type Alias Format
	fmt := new(Alias)
	if err := json.Unmarshal(bytes, fmt); err != nil {
		return err
	}
	*f = Format(*fmt)
	f.Extensions = strings.Split(fmt.FormatName, ",")

	if bitRate, err := strconv.Atoi(fmt.BitRateStr); err == nil {
		f.BitRate = uint(bitRate)
	}
	if size, err := strconv.Atoi(fmt.SizeStr); err == nil {
		f.Size = uint(size)
	}
	if dur, err := time.ParseDuration(fmt.DurationStr + "s"); err == nil {
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
