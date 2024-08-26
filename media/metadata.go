package media

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/graux/goffmpeg"
)

type Metadata struct {
	Streams []Stream `json:"streams"`
	Format  Format   `json:"format"`
}

func (m Metadata) VideoStreams() []Stream {
	return m.filterStreams(CodecTypeVideo)
}

func (m Metadata) AudioStreams() []Stream {
	return m.filterStreams(CodecTypeAudio)
}

func (m Metadata) filterStreams(codecType CodecType) []Stream {
	streams := make([]Stream, 0)
	for _, stream := range m.Streams {
		if stream.CodecType == codecType {
			streams = append(streams, stream)
		}
	}
	return streams
}

func (m Metadata) FirstVideoStream() *Stream {
	return m.firstStream(CodecTypeVideo)
}

func (m Metadata) FirstAudioStream() *Stream {
	return m.firstStream(CodecTypeAudio)
}

func (m Metadata) firstStream(codecType CodecType) *Stream {
	streams := m.filterStreams(codecType)
	if len(streams) == 0 {
		return nil
	}
	return &streams[0]
}

func (m Metadata) IsVideoRotated() *bool {
	videoStream := m.FirstVideoStream()
	if videoStream == nil {
		return nil
	}
	return videoStream.IsRotated()
}

func NewMetadata(cfg goffmpeg.Configuration, inputPath string, whiteListProtocols ...string) (*Metadata, error) {
	var outb, errb bytes.Buffer
	metadata := new(Metadata)
	command := []string{"-i", inputPath, "-print_format", "json", "-show_format", "-show_streams", "-show_error"}

	if len(whiteListProtocols) > 0 {
		command = append([]string{"-protocol_whitelist", strings.Join(whiteListProtocols, ",")}, command...)
	}

	cmd := exec.Command(cfg.FFprobeBinPath(), command...)
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error executing (%s) | error: %s | message: %s %s", command, err, outb.String(), errb.String())
	}

	if err = json.Unmarshal(outb.Bytes(), metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}
