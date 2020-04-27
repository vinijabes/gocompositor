package main

import (
	"log"
	"strings"

	"github.com/vinijabes/gocompositor/gstreamer"
)

func handleMessages(c <-chan *gstreamer.Message) {
	log.Println("Start handling messages")
	for msg := range c {
		log.Println(msg.GetTypeName())
	}
	log.Println("Stop handling messages")
}

func main() {
	gstreamer.ScanPathForPlugins("/usr/lib/x86_64-linux-gnu/gstreamer-1.0/")

	plugins := []string{"videotestsrc", "audiotestsrc", "rtp", "curl", "x264", "rtmp", "audioconvert", "audioresample"}

	err := gstreamer.CheckPlugins(plugins)

	if err != nil {
		panic(err)
	}

	source, err := gstreamer.NewElement("uridecodebin", "source")
	if err != nil {
		log.Fatalln(err)
	}

	convert, err := gstreamer.NewElement("audioconvert", "convert")
	if err != nil {
		log.Fatalln(err)
	}

	resample, err := gstreamer.NewElement("audioresample", "resample")
	if err != nil {
		log.Fatalln(err)
	}

	sink, err := gstreamer.NewElement("fakesink", "sink")
	if err != nil {
		log.Fatalln(err)
	}

	pipeline, err := gstreamer.NewPipeline("teste")
	if err != nil {
		log.Fatalln(err)
	}

	c := pipeline.PullMessage()
	go handleMessages(c)

	pipeline.Add(source)
	pipeline.Add(convert)
	pipeline.Add(resample)
	pipeline.Add(sink)

	err = convert.Link(resample)
	if err != nil {
		log.Fatalln(err)
	}

	err = resample.Link(sink)
	if err != nil {
		log.Fatalln(err)
	}

	source.Set("uri", "https://www.freedesktop.org/software/gstreamer-sdk/data/media/sintel_trailer-480p.webm")
	source.ConnectPadAddedSignal(func(e *gstreamer.Element, pad *gstreamer.Pad) {
		sinkPad := convert.GetStaticPad("sink")
		caps := pad.GetCurrentCaps()
		structure := caps.GetStructure()
		padType := structure.GetName()

		defer caps.Unref()

		if strings.HasPrefix(padType, "audio/x-raw") {
			pad.Link(sinkPad)
		}
	})

	log.Println("Pipeline starting")
	pipeline.Start()

	defer log.Println("Pipeline stoping")
	defer pipeline.Stop()

	for {

	}
}
