package element

import (
	"fmt"

	"github.com/vinijabes/gocompositor/internal/logging"
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
	inputfilter gstreamer.Element
	videodepay  gstreamer.Element
	decodebin   gstreamer.Element
	videoscale  gstreamer.Element
	videofilter gstreamer.Element
	timeoverlay gstreamer.Element
	queue       gstreamer.Element
}

func createInputFilter(codec VideoRTCCodec, id int) (gstreamer.Element, error) {
	var err error
	var caps gstreamer.Caps

	inputfilter, err := gstreamer.NewElement("capsfilter", fmt.Sprintf("inputfilter_%d", id))
	if err != nil {
		return nil, err
	}

	switch codec {
	case VideoRTCCodecVP8:
		caps, err = gstreamer.NewCapsFromString("application/x-rtp, encoding-name=VP8-DRAFT-IETF-01")
		break
	case VideoRTCCodecVP9:
	case VideoRTCCodecH264:
		caps, err = gstreamer.NewCapsFromString("application/x-rtp")
		break
	}

	if err != nil {
		return nil, err
	}

	inputfilter.Set("caps", caps)

	return inputfilter, nil
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

func createDecoder(codec VideoRTCCodec, id int) (gstreamer.Element, error) {
	switch codec {
	case VideoRTCCodecVP8:
		return gstreamer.NewElement("vp8dec", fmt.Sprintf("decoder_%d", id))
	case VideoRTCCodecVP9:
		return gstreamer.NewElement("vp9dec", fmt.Sprintf("decoder_%d", id))
	case VideoRTCCodecH264:
		return gstreamer.NewElement("avdec_h264", fmt.Sprintf("decoder_%d", id))
	}

	return nil, nil
}

func NewVideoRTC(width int, height int, codec VideoRTCCodec) (VideoRTC, error) {
	logging.Debug("creating new RTC video src")
	video := &videoRTC{}

	logging.Debug("creating RTC video appsrc")
	videosrc, err := gstreamer.NewElement("appsrc", fmt.Sprintf("source_%d", videoIDGenerator))
	if err != nil {
		logging.Error(err)
		return nil, err
	}

	videosrc.Set("format", 3)
	videosrc.Set("is-live", true)
	videosrc.Set("do-timestamp", true)

	logging.Debug("creating RTC video input capsfilter")
	inputfilter, err := createInputFilter(codec, videoIDGenerator)
	if err != nil {
		logging.Error(err)
		return nil, err
	}

	logging.Debug("creating RTC video rtp depay")
	videodepay, err := createDepay(codec, videoIDGenerator)
	if err != nil {
		logging.Error(err)
		return nil, err
	}

	logging.Debug("creating RTC video decoder")
	decodebin, err := createDecoder(codec, videoIDGenerator)
	if err != nil {
		logging.Error(err)
		return nil, err
	}

	logging.Debug("creating RTC video scale")
	videoscale, err := gstreamer.NewElement("videoscale", fmt.Sprintf("videoscale_%d", videoIDGenerator))
	if err != nil {
		logging.Error(err)
		return nil, err
	}

	logging.Debug("creating RTC video scale filter")
	videofilter, err := gstreamer.NewElement("capsfilter", fmt.Sprintf("videofilter_%d", videoIDGenerator))
	if err != nil {
		logging.Error(err)
		return nil, err
	}

	logging.Debug("creating RTC video output box")
	videobox, err := gstreamer.NewElement("videobox", fmt.Sprintf("box_%d", videoIDGenerator))
	if err != nil {
		logging.Error(err)
		return nil, err
	}

	logging.Debug("creating RTC video timeoverlay")
	timeoverlay, err := gstreamer.NewElement("timeoverlay", fmt.Sprintf("timeoverlay_%d", videoIDGenerator))
	if err != nil {
		logging.Error(err)
		return nil, err
	}

	logging.Debug("creating RTC video queue")
	queue, err := gstreamer.NewElement("queue", fmt.Sprintf("queue_%d", videoIDGenerator))
	if err != nil {
		logging.Error(err)
		return nil, err
	}

	video.videosrc = videosrc
	video.inputfilter = inputfilter
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
		!pipeline.Add(v.inputfilter) ||
		!pipeline.Add(v.videodepay) ||
		!pipeline.Add(v.decodebin) ||
		!pipeline.Add(v.videoscale) ||
		!pipeline.Add(v.videofilter) ||
		!pipeline.Add(v.timeoverlay) ||
		!pipeline.Add(v.queue) ||
		!pipeline.Add(v.videobox) {
		return ErrVideoSetPipeline
	}

	fmt.Println("Linking video")

	if !v.videosrc.Link(v.inputfilter) ||
		!v.inputfilter.Link(v.videodepay) ||
		!v.videodepay.Link(v.decodebin) ||
		!v.decodebin.Link(v.videoscale) ||
		!v.videoscale.Link(v.videofilter) ||
		!v.videofilter.Link(v.timeoverlay) ||
		!v.timeoverlay.Link(v.queue) ||
		!v.queue.Link(v.videobox) {
		return ErrVideoLinkingSetPipeline
	}

	v.pipeline = pipeline

	return nil
}

func (v *videoRTC) SetSize(width int, height int) {
	logging.Debug(fmt.Sprintf("setting video(%s) size to (%d, %d)", v.videosrc.GetName(), width, height))
	caps, err := gstreamer.NewCapsFromString(fmt.Sprintf("%s,width=%d,height=%d", v.getCapsProps(), width, height))
	if err != nil {
		logging.Error(err)
	}

	v.videofilter.Set("caps", caps)
}

func (v *videoRTC) Push(buffer []byte) error {
	return v.videosrc.Push(buffer)
}
