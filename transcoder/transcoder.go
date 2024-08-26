package transcoder

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"

	"github.com/graux/goffmpeg"
	"github.com/graux/goffmpeg/media"
	"github.com/graux/goffmpeg/pkg/duration"
)

// Transcoder Main struct
type Transcoder struct {
	stdErrPipe         io.ReadCloser
	stdStdinPipe       io.WriteCloser
	process            *exec.Cmd
	mediafile          *media.File
	configuration      goffmpeg.Configuration
	whiteListProtocols []string
}

func NewTranscoder(sourceFile, targetFile string) (*Transcoder, error) {
	tr := new(Transcoder)
	if err := tr.Initialize(sourceFile, targetFile); err != nil {
		return nil, err
	}
	return tr, nil
}

// SetProcessStderrPipe Set the STDERR pipe
func (t *Transcoder) SetProcessStderrPipe(v io.ReadCloser) {
	t.stdErrPipe = v
}

// SetProcessStdinPipe Set the STDIN pipe
func (t *Transcoder) SetProcessStdinPipe(v io.WriteCloser) {
	t.stdStdinPipe = v
}

// SetProcess Set the transcoding process
func (t *Transcoder) SetProcess(cmd *exec.Cmd) {
	t.process = cmd
}

// SetMediaFile Set the media file
func (t *Transcoder) SetMediaFile(v *media.File) {
	t.mediafile = v
}

// SetConfiguration Set the transcoding configuration
func (t *Transcoder) SetConfiguration(v goffmpeg.Configuration) {
	t.configuration = v
}

func (t *Transcoder) SetWhiteListProtocols(availableProtocols []string) {
	t.whiteListProtocols = availableProtocols
}

// Process Get transcoding process
func (t Transcoder) Process() *exec.Cmd {
	return t.process
}

// MediaFile Get the ttranscoding media file.
func (t Transcoder) MediaFile() *media.File {
	return t.mediafile
}

// FFmpegExec Get FFmpeg Bin path
func (t Transcoder) FFmpegExec() string {
	return t.configuration.FFmpegBinPath()
}

// FFprobeExec Get FFprobe Bin path
func (t Transcoder) FFprobeExec() string {
	return t.configuration.FFprobeBinPath()
}

// GetCommand Build and get command
func (t Transcoder) GetCommand() []string {
	media := t.mediafile
	rcommand := append([]string{"-y"}, media.ToStrCommand()...)

	if t.whiteListProtocols != nil {
		rcommand = append([]string{"-protocol_whitelist", strings.Join(t.whiteListProtocols, ",")}, rcommand...)
	}

	return rcommand
}

// InitializeEmptyTranscoder initializes the fields necessary for a blank transcoder
func (t *Transcoder) InitializeEmptyTranscoder() error {
	var err error
	cfg := t.configuration
	if len(cfg.FFmpegBinPath()) == 0 || len(cfg.FFprobeBinPath()) == 0 {
		cfg, err = goffmpeg.Configure(context.Background())
		if err != nil {
			return err
		}
	}
	// Set new File
	MediaFile := new(media.File)
	MediaFile.SetMetadata(new(media.Metadata))

	// Set transcoder configuration
	t.SetMediaFile(MediaFile)
	t.SetConfiguration(cfg)
	return nil
}

// SetInputPath sets the input path for transcoding
func (t *Transcoder) SetInputPath(inputPath string) error {
	if t.mediafile.InputPipe() {
		return errors.New("cannot set an input path when an input pipe command has been set")
	}
	t.mediafile.SetInputPath(inputPath)
	return nil
}

// SetOutputPath sets the output path for transcoding
func (t *Transcoder) SetOutputPath(inputPath string) error {
	if t.mediafile.OutputPipe() {
		return errors.New("cannot set an input path when an input pipe command has been set")
	}
	t.mediafile.SetOutputPath(inputPath)
	return nil
}

// CreateInputPipe creates an input pipe for the transcoding process
func (t *Transcoder) CreateInputPipe() (*io.PipeWriter, error) {
	if t.mediafile.InputPath() != "" {
		return nil, errors.New("cannot set an input pipe when an input path exists")
	}
	inputPipeReader, inputPipeWriter := io.Pipe()
	t.mediafile.SetInputPipe(true)
	t.mediafile.SetInputPipeReader(inputPipeReader)
	t.mediafile.SetInputPipeWriter(inputPipeWriter)
	return inputPipeWriter, nil
}

// CreateOutputPipe creates an output pipe for the transcoding process
func (t *Transcoder) CreateOutputPipe(containerFormat string) (*io.PipeReader, error) {
	if t.mediafile.OutputPath() != "" {
		return nil, errors.New("cannot set an output pipe when an output path exists")
	}
	t.mediafile.SetOutputFormat(containerFormat)

	t.mediafile.SetMovFlags("frag_keyframe")
	outputPipeReader, outputPipeWriter := io.Pipe()
	t.mediafile.SetOutputPipe(true)
	t.mediafile.SetOutputPipeReader(outputPipeReader)
	t.mediafile.SetOutputPipeWriter(outputPipeWriter)
	return outputPipeReader, nil
}

// Initialize Init the transcoding process
func (t *Transcoder) Initialize(inputPath string, outputPath string) error {
	var err error
	cfg := t.configuration
	if len(cfg.FFmpegBinPath()) == 0 || len(cfg.FFprobeBinPath()) == 0 {
		cfg, err = goffmpeg.Configure(context.Background())
		if err != nil {
			return err
		}
	}

	if inputPath == "" {
		return errors.New("error on transcoder.Initialize: inputPath missing")
	}

	metadata, err := media.NewMetadata(cfg, inputPath, t.whiteListProtocols...)
	if err != nil {
		return err
	}

	// Set new File
	MediaFile := new(media.File)
	MediaFile.SetMetadata(metadata)
	MediaFile.SetInputPath(inputPath)
	MediaFile.SetOutputPath(outputPath)

	// Set transcoder configuration
	t.SetMediaFile(MediaFile)
	t.SetConfiguration(cfg)

	return nil
}

// Run Starts the transcoding process
func (t *Transcoder) Run(progress bool) <-chan error {
	done := make(chan error)
	command := t.GetCommand()

	if !progress {
		command = append([]string{"-nostats", "-loglevel", "0"}, command...)
	}

	proc := exec.Command(t.configuration.FFmpegBinPath(), command...)
	if progress {
		errStream, err := proc.StderrPipe()
		if err != nil {
			fmt.Println("Progress not available: " + err.Error())
		} else {
			t.stdErrPipe = errStream
		}
	}

	// Set the stdinPipe in case we need to stop the transcoding
	stdin, err := proc.StdinPipe()
	if nil != err {
		fmt.Println("Stdin not available: " + err.Error())
	}

	t.stdStdinPipe = stdin

	// If the user has requested progress, we send it to them on a Buffer
	var outb, errb bytes.Buffer
	if progress {
		proc.Stdout = &outb
	}

	// If an input pipe has been set, we set it as stdin for the transcoding
	if t.mediafile.InputPipe() {
		proc.Stdin = t.mediafile.InputPipeReader()
	}

	// If an output pipe has been set, we set it as stdout for the transcoding
	if t.mediafile.OutputPipe() {
		proc.Stdout = t.mediafile.OutputPipeWriter()
	}

	err = proc.Start()

	t.SetProcess(proc)

	go func(err error) {
		if err != nil {
			done <- fmt.Errorf("failed start ffmpeg (%s) with %s, message %s %s", command, err, outb.String(), errb.String())
			close(done)
			return
		}

		err = proc.Wait()
		if err != nil {
			err = fmt.Errorf("failed finish ffmpeg (%s) with %s message %s %s", command, err, outb.String(), errb.String())
		}

		go t.closePipes()

		done <- err
		close(done)
	}(err)

	return done
}

// Stop Ends the transcoding process
func (t *Transcoder) Stop() error {
	if t.process != nil {
		stdin := t.stdStdinPipe
		if stdin != nil {
			if _, err := stdin.Write([]byte("q\n")); err != nil {
				return err
			}
		}
	}
	return nil
}

// Output Returns the transcoding progress channel
func (t Transcoder) Output() <-chan Progress {
	out := make(chan Progress)

	go func() {
		defer close(out)
		if t.stdErrPipe == nil {
			out <- Progress{}
			return
		}

		defer t.stdErrPipe.Close()

		scanner := bufio.NewScanner(t.stdErrPipe)

		split := func(data []byte, atEOF bool) (advance int, token []byte, spliterror error) {
			if atEOF && len(data) == 0 {
				return 0, nil, nil
			}
			// windows \r\n
			// so  first \r and then \n can remove unexpected line break
			if i := bytes.IndexByte(data, '\r'); i >= 0 {
				// We have a cr terminated line
				return i + 1, data[0:i], nil
			}
			if i := bytes.IndexByte(data, '\n'); i >= 0 {
				// We have a full newline-terminated line.
				return i + 1, data[0:i], nil
			}
			if atEOF {
				return len(data), data, nil
			}

			return 0, nil, nil
		}

		scanner.Split(split)
		buf := make([]byte, 2)
		scanner.Buffer(buf, bufio.MaxScanTokenSize)

		for scanner.Scan() {
			Progress := new(Progress)
			line := scanner.Text()
			if strings.Contains(line, "frame=") && strings.Contains(line, "time=") && strings.Contains(line, "bitrate=") {
				re := regexp.MustCompile(`=\s+`)
				st := re.ReplaceAllString(line, `=`)

				f := strings.Fields(st)
				var framesProcessed string
				var currentTime string
				var currentBitrate string
				var currentSpeed string

				for j := 0; j < len(f); j++ {
					field := f[j]
					fieldSplit := strings.Split(field, "=")

					if len(fieldSplit) > 1 {
						fieldname := strings.Split(field, "=")[0]
						fieldvalue := strings.Split(field, "=")[1]

						if fieldname == "frame" {
							framesProcessed = fieldvalue
						}

						if fieldname == "time" {
							currentTime = fieldvalue
						}

						if fieldname == "bitrate" {
							currentBitrate = fieldvalue
						}
						if fieldname == "speed" {
							currentSpeed = fieldvalue
						}
					}
				}

				timesec := duration.DurToSec(currentTime)
				dur := t.MediaFile().Metadata().Format.Duration
				// live stream check
				if dur > 0 {
					// Progress calculation
					progress := (timesec * 100) / dur.Seconds()
					Progress.Progress = progress
				}
				Progress.CurrentBitrate = currentBitrate
				Progress.FramesProcessed = framesProcessed
				Progress.CurrentTime = currentTime
				Progress.Speed = currentSpeed
				out <- *Progress
			}
		}
	}()

	return out
}

func (t *Transcoder) closePipes() {
	if t.mediafile.InputPipe() {
		t.mediafile.InputPipeReader().Close()
	}
	if t.mediafile.OutputPipe() {
		t.mediafile.OutputPipeWriter().Close()
	}
}
