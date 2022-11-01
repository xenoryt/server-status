package main

import (
	"embed"
	"fmt"
	"log"
	"os"

	"github.com/DavidGamba/go-getoptions"
)

//go:embed client/dist/*
var client embed.FS

const HLSStaticURL = "https://rtmp.skywardbox.net/static"

var (
	logger   *log.Logger
	mediaDir *string
	//streamUrl    *string

	streamOutDir *string
)

func main() {
	logger = log.New(os.Stdout, "", 0)

	opts := getoptions.New()
	mediaDir = opts.String("media-dir", "", opts.Alias("d"),
		opts.Description("Directory of media files to stream"),
		opts.GetEnv("MEDIA_DIR"),
		opts.Required(),
	)
	//streamUrl = opts.String("stream-url", "", opts.Alias("s"),
	//	opts.Description("URL to send the stream to."),
	//	opts.GetEnv("STREAM_URL"),
	//	//opts.Required(),
	//)
	streamOutDir = opts.String("output-dir", "",
		opts.Description("Directory the HLS files should be output to"),
		opts.GetEnv("OUTPUT_DIR"),
		opts.Required(),
	)

	_, err := opts.Parse(os.Args[1:])

	if err != nil {
		fmt.Println(opts.Help())
		os.Exit(1)
	}

	StartServer()
}
