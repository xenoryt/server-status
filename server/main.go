package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DavidGamba/go-getoptions"
	"github.com/xenoryt/skv"
	"github.com/xenoryt/stream-status/lib/stream"
)

//go:embed client/dist/*
var client embed.FS

var db *skv.KVStore

const HLSStaticURL = "https://rtmp.skywardbox.net/static"

var (
	logger   *log.Logger
	mediaDir *string
	//streamUrl    *string

	streamOutDir *string
)

func main() {
	logger = log.New(os.Stdout, "", 0)

	var err error
	db, err = skv.Open("streamSession.db")
	if err != nil {
		logger.Fatalln("Failed to open session db:", err)
	}
	var streams []*stream.Stream
	if err = db.Get("streams", &streams); err != skv.ErrNotFound {
		logger.Fatalln("Failed to read db:", err)
	}
	logger.Printf("Loaded streams: %+q\n", streams)
	stream.LoadStreams(streams)

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

	_, err = opts.Parse(os.Args[1:])

	if err != nil {
		fmt.Println(opts.Help())
		os.Exit(1)
	}

	go handleSignals()

	StartServer()
}

func handleSignals() {
	var captureSignal = make(chan os.Signal, 1)
	signal.Notify(captureSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)

	sig := <-captureSignal
	logger.Println("Recieved signal:", sig)

	logger.Println("Writing to db...")
	db.Put("streams", stream.ListStreams())
	db.Close()

	logger.Println("Wait for 1 second to finish processing")
	time.Sleep(time.Second)
	os.Exit(0)
}
