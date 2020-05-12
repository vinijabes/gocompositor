package main

import (
	"log"
	"time"

	"github.com/vinijabes/gocompositor/compositor"
	gstreamer "github.com/vinijabes/gostreamer"
)

func handleMessages(c <-chan *gstreamer.Message) {
	log.Println("Start handling messages")
	for msg := range c {
		log.Println(msg.GetTypeName())
	}
	log.Println("Stop handling messages")
}

func main() {
	cmp, err := compositor.NewCompositor()
	if err != nil {
		log.Fatalln(err)
	}

	video, err := compositor.NewVideo(640, 360, "videotestsrc")
	if err != nil {
		log.Fatalln(err)
	}

	video2, err := compositor.NewTestVideo(640, 360, 23)
	if err != nil {
		log.Fatalln(err)
	}

	video3, err := compositor.NewTestVideo(640, 360, 1)
	if err != nil {
		log.Fatalln(err)
	}

	video4, err := compositor.NewTestVideo(640, 360, 24)
	if err != nil {
		log.Fatalln(err)
	}

	err = cmp.AddVideo(video)
	if err != nil {
		log.Fatalln(err)
	}

	convert, err := gstreamer.NewElement("videoconvert", "convert")
	if err != nil {
		log.Fatalln(err)
	}

	enc, err := gstreamer.NewElement("x264enc", "enc")
	if err != nil {
		log.Fatalln(err)
	}

	mux, err := gstreamer.NewElement("flvmux", "mux")
	if err != nil {
		log.Fatalln(err)
	}

	sink, err := gstreamer.NewElement("rtmpsink", "sink")
	if err != nil {
		log.Fatalln(err)
	}

	cmp.Add(convert)
	cmp.Add(enc)
	cmp.Add(mux)
	cmp.Add(sink)

	cmp.LinkVideoSink(convert)
	convert.Link(enc)
	enc.Link(mux)
	mux.Link(sink)

	sink.Set("location", "rtmp://teste.com")

	layout := compositor.NewLayout(1280, 720)
	videoRule1 := compositor.NewLayoutRule()
	slot1 := compositor.NewLayoutSlotWithSymetricBorders(0, 0, 640, 360, 320, 180)

	videoRule1.AddSlot(slot1)

	videoRule2 := compositor.NewLayoutRule()
	slot2 := compositor.NewLayoutSlotWithSymetricBorders(0, 0, 640, 360, 0, 180)
	slot3 := compositor.NewLayoutSlotWithSymetricBorders(640, 0, 640, 360, 0, 180)

	videoRule2.AddSlot(slot2)
	videoRule2.AddSlot(slot3)

	videoRule3 := compositor.NewLayoutRule()
	slot4 := compositor.NewLayoutSlot(0, 0, 640, 360)
	slot5 := compositor.NewLayoutSlot(640, 0, 640, 360)
	slot6 := compositor.NewLayoutSlot(320, 360, 640, 360)

	videoRule3.AddSlot(slot4)
	videoRule3.AddSlot(slot5)
	videoRule3.AddSlot(slot6)

	videoRule4 := compositor.NewLayoutRule()
	slot7 := compositor.NewLayoutSlot(0, 0, 640, 360)
	slot8 := compositor.NewLayoutSlot(640, 0, 640, 360)
	slot9 := compositor.NewLayoutSlot(0, 360, 640, 360)
	slot10 := compositor.NewLayoutSlot(640, 360, 640, 360)

	videoRule4.AddSlot(slot7)
	videoRule4.AddSlot(slot8)
	videoRule4.AddSlot(slot9)
	videoRule4.AddSlot(slot10)

	layout.AddRule(videoRule1, 1)
	layout.AddRule(videoRule2, 2)
	layout.AddRule(videoRule3, 3)
	layout.AddRule(videoRule4, 4)
	cmp.SetLayout(layout)

	// box2.SetPos(int64(640), int64(0))
	// box3.SetPos(int64(0), int64(360))
	// box4.SetPos(int64(640), int64(360))

	cmp.Start()
	defer cmp.Stop()

	time.Sleep(20 * time.Second)

	cmp.Pause()
	err = cmp.AddVideo(video2)
	if err != nil {
		log.Fatalln(err)
	}
	cmp.Start()

	time.Sleep(20 * time.Second)

	cmp.Pause()
	err = cmp.AddVideo(video3)
	if err != nil {
		log.Fatalln(err)
	}

	cmp.Start()
	time.Sleep(20 * time.Second)

	cmp.Pause()
	err = cmp.AddVideo(video4)
	if err != nil {
		log.Fatalln(err)
	}

	cmp.Start()

	for {

	}
}
