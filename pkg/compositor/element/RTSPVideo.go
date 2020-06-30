package element

import (
	"fmt"

	"github.com/vinijabes/gostreamer/pkg/gstreamer"
)

type VideoRTSP interface {
	Video
}

type videoRTSP struct {
	video
	decodebin   gstreamer.Element
	videoscale  gstreamer.Element
	videofilter gstreamer.Element
	timeoverlay gstreamer.Element
	queue       gstreamer.Element
}

func NewVideoRTSP(width int, height int, location string, latency int) (VideoRTSP, error) {
	video := &videoRTSP{}
	videosrc, err := gstreamer.NewElement("rtspsrc", fmt.Sprintf("source_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	videosrc.Set("location", location)
	videosrc.Set("latency", latency)

	decodebin, err := gstreamer.NewElement("decodebin", fmt.Sprintf("decodebin_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	videoscale, err := gstreamer.NewElement("videoscale", fmt.Sprintf("videoscale_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	videofilter, err := gstreamer.NewElement("capsfilter", fmt.Sprintf("videofilter_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	videobox, err := gstreamer.NewElement("videobox", fmt.Sprintf("box_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	timeoverlay, err := gstreamer.NewElement("timeoverlay", fmt.Sprintf("timeoverlay_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	queue, err := gstreamer.NewElement("queue", fmt.Sprintf("queue_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	video.videosrc = videosrc
	video.decodebin = decodebin
	video.videoscale = videoscale
	video.videofilter = videofilter
	video.timeoverlay = timeoverlay
	video.queue = queue
	video.videobox = videobox

	videoIDGenerator++
	video.SetSize(width, height)

	return video, nil
}

func (v *videoRTSP) SetPipeline(pipeline gstreamer.Pipeline) error {
	if v.pipeline != nil {
		v.videosrc.Unlink(v.videofilter)
		v.videofilter.Unlink(v.videobox)

		if !v.pipeline.Remove(v.videosrc) ||
			!v.pipeline.Remove(v.videofilter) ||
			!v.pipeline.Remove(v.videobox) {
			return ErrVideoSetPipeline
		}
	}

	if !pipeline.Add(v.videosrc) ||
		!pipeline.Add(v.decodebin) ||
		!pipeline.Add(v.videoscale) ||
		!pipeline.Add(v.videofilter) ||
		!pipeline.Add(v.timeoverlay) ||
		!pipeline.Add(v.queue) ||
		!pipeline.Add(v.videobox) {
		return ErrVideoSetPipeline
	}

	v.videosrc.SetOnPadAddedCallback(func(element gstreamer.Element, pad gstreamer.Pad) {
		fmt.Println("VideoSrc pad-added")
		sinkpad, err := v.decodebin.GetStaticPad("sink")
		if err != nil {
			fmt.Println(err)
		}

		result := pad.Link(sinkpad)
		if result != gstreamer.GstPadLinkOk {
			fmt.Println("Failed to link rtsp pad")
		}
	})

	v.decodebin.SetOnPadAddedCallback(func(element gstreamer.Element, pad gstreamer.Pad) {
		fmt.Println("Decodebin pad-added")
		sinkpad, err := v.videoscale.GetStaticPad("sink")
		if err != nil {
			fmt.Println(err)
		}

		result := pad.Link(sinkpad)
		if result != gstreamer.GstPadLinkOk {
			fmt.Println("Failed to link rtsp pad")
		}
	})

	if !v.videoscale.Link(v.videofilter) ||
		!v.videofilter.Link(v.timeoverlay) ||
		!v.timeoverlay.Link(v.videobox) {
		return ErrVideoLinkingSetPipeline
	}

	v.pipeline = pipeline

	return nil
}

func (v *videoRTSP) SetSize(width int, height int) {
	caps, _ := gstreamer.NewCapsFromString(fmt.Sprintf("%s,width=%d,height=%d", v.getCapsProps(), width, height))
	v.videofilter.Set("caps", caps)
}
