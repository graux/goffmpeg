package media

type Metadata struct {
	Streams []Streams `json:"streams"`
	Format  Format    `json:"format"`
}

type Format struct {
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

type Tags struct {
	Encoder string `json:"ENCODER"`
}

func (m Metadata) VideoStreams() []Streams {
	return m.filterStreams(CodecTypeVideo)
}

func (m Metadata) AudioStreams() []Streams {
	return m.filterStreams(CodecTypeAudio)
}

func (m Metadata) filterStreams(codecType CodecType) []Streams {
	streams := make([]Streams, 0)
	for _, stream := range m.Streams {
		if stream.CodecType == codecType {
			streams = append(streams, stream)
		}
	}
	return streams
}

func (m Metadata) FirstVideoStream() *Streams {
	return m.firstStream(CodecTypeVideo)
}

func (m Metadata) FirstAudioStream() *Streams {
	return m.firstStream(CodecTypeAudio)
}

func (m Metadata) firstStream(codecType CodecType) *Streams {
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
