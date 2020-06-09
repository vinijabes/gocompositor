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
	videofilter  gstreamer.Element
	videodepay   gstreamer.Element
	videodecoder gstreamer.Element
	videoscale   gstreamer.Element
}

func NewVideoRTSP(width int, height int, location string) (VideoRTSP, error) {
	video := &videoRTSP{}
	videosrc, err := gstreamer.NewElement("rtspsrc", fmt.Sprintf("source_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	videosrc.Set("location", location)
	videosrc.Set("latency", 0)

	videodepay, err := gstreamer.NewElement("rtph264depay", fmt.Sprintf("videodepay_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	videodecoder, err := gstreamer.NewElement("avdec_h264", fmt.Sprintf("videodecoder_%d", videoIDGenerator))
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

	video.videosrc = videosrc
	video.videofilter = videofilter
	video.videodepay = videodepay
	video.videodecoder = videodecoder
	video.videoscale = videoscale
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
		!pipeline.Add(v.videofilter) ||
		!pipeline.Add(v.videodepay) ||
		!pipeline.Add(v.videodecoder) ||
		!pipeline.Add(v.videoscale) ||
		!pipeline.Add(v.videobox) {
		return ErrVideoSetPipeline
	}

	v.videosrc.SetOnPadAddedCallback(func(element gstreamer.Element, pad gstreamer.Pad) {
		sinkpad, err := v.videodepay.GetStaticPad("sink")
		if err != nil {
			fmt.Println(err)
		}

		result := pad.Link(sinkpad)
		if result != gstreamer.GstPadLinkOk {
			fmt.Println("Failed to link rtsp pad")
		}
	})

	if !v.videodepay.Link(v.videodecoder) ||
		!v.videodecoder.Link(v.videoscale) ||
		!v.videoscale.Link(v.videofilter) ||
		!v.videofilter.Link(v.videobox) {
		return ErrVideoLinkingSetPipeline
	}

	v.pipeline = pipeline

	return nil
}

func (v *videoRTSP) SetSize(width int, height int) {
	caps, _ := gstreamer.NewCapsFromString(fmt.Sprintf("%s,width=%d,height=%d", v.getCapsProps(), width, height))
	v.videofilter.Set("caps", caps)
}
