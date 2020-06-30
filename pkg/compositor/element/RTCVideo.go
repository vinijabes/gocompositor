package element

import (
	"fmt"

	"github.com/vinijabes/gostreamer/pkg/gstreamer"
)

type VideoRTCCodec string

const (
	VideoRTCCodecVP8  VideoRTCCodec = "VP8"
	VideoRTCCodecVP9  VideoRTCCodec = "VP9"
	VideoRTCCodecH264 VideoRTCCodec = "H264"
	VideoRTCCodecOpus VideoRTCCodec = "opus"
	VideoRTCCodecG722 VideoRTCCodec = "G722"
	VideoRTCCodecPCMA VideoRTCCodec = "PCMA"
	VideoRTCCodecPCMU VideoRTCCodec = "PCMU"
)

type VideoRTC interface {
	Video
	Push(buffer []byte) error
}

type videoRTC struct {
	video
	videodepay  gstreamer.Element
	decodebin   gstreamer.Element
	videoscale  gstreamer.Element
	videofilter gstreamer.Element
	timeoverlay gstreamer.Element
	queue       gstreamer.Element
}

func createDepay(codec VideoRTCCodec, id int) (gstreamer.Element, error) {
	switch codec {
	case VideoRTCCodecVP8:
		return gstreamer.NewElement("rtpvp8depay", fmt.Sprintf("depay_%d", id))
	case VideoRTCCodecVP9:
		return gstreamer.NewElement("rtpvp9depay", fmt.Sprintf("depay_%d", id))
	case VideoRTCCodecH264:
		return gstreamer.NewElement("rtph264depay", fmt.Sprintf("depay_%d", id))
	}

	return nil, nil
}

func NewVideoRTC(width int, height int, codec VideoRTCCodec) (VideoRTC, error) {
	video := &videoRTC{}
	videosrc, err := gstreamer.NewElement("appsrc", fmt.Sprintf("source_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	videosrc.Set("format", 3)
	videosrc.Set("is-live", true)
	videosrc.Set("do-timestamp", true)

	videodepay, err := createDepay(codec, videoIDGenerator)
	if err != nil {
		return nil, err
	}

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
	video.videodepay = videodepay
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

func (v *videoRTC) SetPipeline(pipeline gstreamer.Pipeline) error {
	if v.pipeline != nil {
		v.videosrc.Unlink(v.videofilter)
		v.videofilter.Unlink(v.videobox)

		if !v.pipeline.Remove(v.videosrc) ||
			!v.pipeline.Remove(v.decodebin) ||
			!v.pipeline.Remove(v.videoscale) ||
			!v.pipeline.Remove(v.videofilter) ||
			!v.pipeline.Remove(v.timeoverlay) ||
			!v.pipeline.Remove(v.queue) ||
			!v.pipeline.Remove(v.videobox) {
			return ErrVideoSetPipeline
		}
	}

	if !pipeline.Add(v.videosrc) ||
		!pipeline.Add(v.videodepay) ||
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

func (v *videoRTC) SetSize(width int, height int) {
	caps, _ := gstreamer.NewCapsFromString(fmt.Sprintf("%s,width=%d,height=%d", v.getCapsProps(), width, height))
	v.videofilter.Set("caps", caps)
}

func (v *videoRTC) Push(buffer []byte) error {
	return v.videosrc.Push(buffer)
}
