package stream

import (
	"fmt"
	"os"
	"os/exec"
	"path"
)

var ErrStreamExists = fmt.Errorf("Another stream is already running")

type Stream struct {
	Cmd  *exec.Cmd `json:"-"`
	Done bool      `json:"done"`

	Filepath   string `json:"filepath"`
	OutputPath string `json:"outputPath"`

	Url string `json:"url"`
}

func (s Stream) Status() string {
	if s.Done {
		return "done"
	} else if s.Cmd == nil {
		return "cancelled"
	}
	return "active"
}

func (s Stream) String() string {
	return fmt.Sprintf("'...%s' - %s", s.Filepath[len(s.Filepath)-10:], s.Status())
}

var streams []*Stream

func LoadStreams(s []*Stream) {
	streams = s
}

func streamCommand(outputFilepath, filepath string, subs bool) []string {
	subsOpts := []string{
		"-filter_complex", fmt.Sprintf("subtitles='%s'", filepath),
	}

	// "ffmpeg", "-re", "-i", filepath, "-c:v", "libx264", "-c:a", "aac", "-f", "flv",
	command := []string{
		"ffmpeg", "-i", filepath, "-s", "1920x1080", "-map", "0:1", "-c:a", "aac", "-ar", "48000", "-b:a", "128k", "-ac", "2", "-map", "0:0", "-hls_time", "5", "-hls_list_size", "0", "-f", "hls",
	}

	if subs {
		command = append(command, subsOpts...)
	}

	command = append(command, outputFilepath)
	return command
}

// StreamFile streams a media file to a url.
// If subs is true, will add hard subs to the stream.
func StreamFile(outputFilepath, filepath string, subs bool) error {
	s := GetStream(filepath)

	if s != nil && !s.Done {
		return ErrStreamExists
	}

	if err := os.MkdirAll(path.Dir(outputFilepath), 0775); err != nil {
		return err
	}

	command := streamCommand(outputFilepath, filepath, subs)
	streamCmd := exec.Command(command[0], command[1:]...)
	streamCmd.Stdout = os.Stdout
	streamCmd.Stderr = os.Stderr
	err := streamCmd.Start()
	if err != nil {
		return err
	}
	stream := Stream{
		Cmd:        streamCmd,
		Filepath:   filepath,
		OutputPath: outputFilepath,
	}
	streams = append(streams, &stream)

	go func() {
		stream.Cmd.Wait()
		stream.Done = true
	}()

	return nil
}

// StopStream stops any ongoing streams.
func StopStream(filepath string) error {
	s := GetStream(filepath)
	if s == nil {
		return nil
	}
	return s.Cmd.Process.Kill()
}

func GetStream(filepath string) *Stream {
	for _, s := range streams {
		if s.Filepath == filepath {
			return s
		}
	}
	return nil
}

func ListStreams() []*Stream {
	return streams
}
