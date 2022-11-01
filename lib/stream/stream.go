package stream

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"sync"
)

var ErrStreamExists = fmt.Errorf("Another stream is already running")

var streamCmd *exec.Cmd
var streamLock sync.Mutex
var streaming bool

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

// Streaming returns true iff there is an active stream.
func Streaming() bool {
	return streaming
}

// StreamFile streams a media file to a url.
// If subs is true, will add hard subs to the stream.
func StreamFile(outputFilepath, filepath string, subs bool) error {
	if !streamLock.TryLock() {
		return ErrStreamExists
	}
	defer streamLock.Unlock()

	if streaming {
		return ErrStreamExists
	}

	if err := os.MkdirAll(path.Dir(outputFilepath), 0775); err != nil {
		return err
	}

	command := streamCommand(outputFilepath, filepath, subs)
	streamCmd = exec.Command(command[0], command[1:]...)
	streamCmd.Stdout = os.Stdout
	streamCmd.Stderr = os.Stderr
	err := streamCmd.Start()
	if err != nil {
		return err
	}

	streaming = true
	go func() {
		streamCmd.Wait()
		streamLock.Lock()
		streaming = false
		streamLock.Unlock()
	}()

	return nil
}

// StopStream stops any ongoing streams.
func StopStream() error {
	if streamCmd == nil {
		return nil
	}
	return streamCmd.Process.Kill()
}
