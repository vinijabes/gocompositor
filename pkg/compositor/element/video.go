package element

import (
	"errors"
	"fmt"

	"github.com/vinijabes/gostreamer/pkg/gstreamer"
)

type Video interface {
	SetPos(x int, y int)
	SetSize(width int, height int)
	SetBorder(border VideoBorder, value int)
	SetPipeline(pipeline gstreamer.Pipeline) error

	LinkSinkPad(gstreamer.Pad) (gstreamer.GstPadLinkReturn, error)

	Raw() gstreamer.Element
}

type Videos []Video

type video struct {
	videosrc  gstreamer.Element
	videobox  gstreamer.Element
	videosink gstreamer.Pad

	pipeline gstreamer.Pipeline
}

//VideoBorder ...
type VideoBorder int

//Border direction constants
const (
	VideoBorderTop VideoBorder = 1 << iota
	VideoBorderRight
	VideoBorderBottom
	VideoBorderLeft
)

var (
	ErrVideoSetPipeline        = errors.New("Failed to set video pipeline")
	ErrVideoLinkingSetPipeline = errors.New("Failed to link elements when setting video pipeline")
)

var videoIDGenerator = 0

//NewVideo returns a gstreamer video wrapper
func NewVideo(width int, height int, factory string) (Video, error) {
	video := &video{}
	videosrc, err := gstreamer.NewElement(factory, fmt.Sprintf("source_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	videobox, err := gstreamer.NewElement("videobox", fmt.Sprintf("box_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	video.videosrc = videosrc
	video.videobox = videobox

	videoIDGenerator++
	video.SetSize(width, height)

	return video, nil
}

func (v *video) SetPos(x int, y int) {
	if v.videosink != nil {
		v.videosink.Set("xpos", x)
		v.videosink.Set("ypos", y)
	}
}

func (v *video) SetSize(width int, height int) {
	caps, _ := gstreamer.NewCapsFromString(fmt.Sprintf("%s,width=%d,height=%d", v.getCapsProps(), width, height))
	v.videosrc.Set("caps", caps)
}

func (v *video) SetBorder(border VideoBorder, value int) {
	if border&VideoBorderTop != 0 {
		v.videobox.Set("top", value)
	}

	if border&VideoBorderRight != 0 {
		v.videobox.Set("right", value)
	}

	if border&VideoBorderBottom != 0 {
		v.videobox.Set("bottom", value)
	}

	if border&VideoBorderLeft != 0 {
		v.videobox.Set("left", value)
	}
}

func (v *video) SetPipeline(pipeline gstreamer.Pipeline) error {
	if v.pipeline != nil {
		v.videosrc.Unlink(v.videobox)

		if !v.pipeline.Remove(v.videosrc) || !v.pipeline.Remove(v.videobox) {
			return ErrVideoSetPipeline
		}
	}

	if !pipeline.Add(v.videobox) || !pipeline.Add(v.videosrc) || !v.videosrc.Link(v.videobox) {
		return ErrVideoSetPipeline
	}

	v.pipeline = pipeline

	return nil
}

func (v *video) LinkSinkPad(sink gstreamer.Pad) (gstreamer.GstPadLinkReturn, error) {
	srcpad, err := v.videobox.GetStaticPad("src")
	if err != nil {
		return gstreamer.GstPadLinkRefused, err
	}

	result := srcpad.Link(sink)

	if result == gstreamer.GstPadLinkOk {
		v.videosink = sink
	}

	return result, nil
}

func (v *video) Raw() gstreamer.Element {
	return v.videosrc
}

func (v *video) getCapsProps() string {
	return "video/x-raw"
}
