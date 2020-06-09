package compositor

import (
	"fmt"

	gstreamer "github.com/vinijabes/gostreamer/pkg/gstreamer"
)

type box struct {
	gstSink gstreamer.Pad
}

type Video interface {
	SetPos(x int64, y int64)
	SetSize(width int64, height int64)
	SetBorder(border string, value int64)
	SetPipeline(pipeline gstreamer.Pipeline) error
	SetBox(b *box)

	GetSrcPad() (gstreamer.Pad, error)
}

type RTCVideo interface {
	Video

	Push(buffer []byte)
}

//Video ...
type video struct {
	pipeline    gstreamer.Pipeline
	gstFilter   gstreamer.Element
	gstVideobox gstreamer.Element
	gstSrc      gstreamer.Element
	mediaType   string
	id          int

	box *box
}

type Codec int

const (
	CodecVP8 = iota + 1
	CodecVP9
	CodecH264
)

type rtcVideo struct {
	video
	gstDepay gstreamer.Element
	codec    Codec
}

//Videos ...
type Videos []Video

var videoIDGenerator = 10

//NewRawVideo ...
func NewRawVideo() Video {
	return &video{}
}

//NewTestVideo ...
func NewTestVideo(width int, height int, pattern int) (Video, error) {
	v, err := NewVideo(width, height, "videotestsrc")
	if err != nil {
		return nil, err
	}

	raw := v.(*video)
	raw.gstSrc.Set("pattern", pattern)

	return v, nil
}

//NewVideo ...
func NewVideo(width int, height int, srcPlugin string) (Video, error) {
	src, err := gstreamer.NewElement(srcPlugin, fmt.Sprintf("source_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	return NewVideoFromElement(src, width, height)
}

//NewVideoFromElement ...
func NewVideoFromElement(src gstreamer.Element, width int, height int) (Video, error) {
	filter, err := gstreamer.NewElement("capsfilter", fmt.Sprintf("filter_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	videobox, err := gstreamer.NewElement("videobox", fmt.Sprintf("box_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	video := &video{
		gstSrc:      src,
		gstFilter:   filter,
		gstVideobox: videobox,
		mediaType:   "video/x-raw",
		id:          videoIDGenerator,
	}

	video.SetSize(int64(width), int64(height))

	videoIDGenerator++

	return video, nil
}

func getDepayFromCodec(codec Codec) string {
	if codec == CodecVP8 {
		return "rtpvp8depay"
	} else if codec == CodecVP9 {
		return "rtpvp9depay"
	} else if codec == CodecH264 {
		return "rtph264depay"
	}

	panic("Unknown codec")
}

//NewRTCVideo ...
func NewRTCVideo(codec Codec, width int, height int) (RTCVideo, error) {
	src, err := gstreamer.NewElement("appsrc", fmt.Sprintf("source_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}
	src.Set("format", 2)
	src.Set("is-live", true)
	src.Set("do-timestamp", true)

	filter, err := gstreamer.NewElement("capsfilter", fmt.Sprintf("filter_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	depay, err := gstreamer.NewElement(getDepayFromCodec(codec), fmt.Sprintf("depay_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	videobox, err := gstreamer.NewElement("videobox", fmt.Sprintf("box_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	v := &rtcVideo{
		video: video{
			gstSrc:      src,
			gstFilter:   filter,
			gstVideobox: videobox,
			mediaType:   "video/x-raw",
			id:          videoIDGenerator,
		},
		codec:    codec,
		gstDepay: depay,
	}

	v.SetSize(int64(width), int64(height))

	return v, nil
}

//LinkSrc link external element with src element
func (v *video) LinkSrc(e gstreamer.Element) bool {
	return e.Link(v.gstSrc)
}

func (v *video) internalLink() error {
	res := v.gstSrc.Link(v.gstFilter)
	if !res {
		return fmt.Errorf("failed to link src with filter")
	}

	res = v.gstFilter.Link(v.gstVideobox)
	if !res {
		return fmt.Errorf("Failed to link filter with box")
	}

	return nil
}

//SetPos ...
func (v *video) SetPos(x int64, y int64) {
	v.box.gstSink.Set("xpos", x)
	v.box.gstSink.Set("ypos", y)
}

//SetSize ...
func (v *video) SetSize(width int64, height int64) {
	caps, _ := gstreamer.NewCapsFromString(fmt.Sprintf("%s,width=%d,height=%d", v.mediaType, width, height))
	v.gstFilter.Set("caps", caps)
}

//SetBorder ...
func (v *video) SetBorder(border string, value int64) {
	v.gstVideobox.Set(border, value)
}

func (v *video) SetPipeline(pipeline gstreamer.Pipeline) error {
	pipeline.Add(v.gstSrc)
	pipeline.Add(v.gstFilter)
	pipeline.Add(v.gstVideobox)

	return v.internalLink()
}

func (v *video) SetBox(b *box) {
	v.box = b
}

func (v *video) GetSrcPad() (gstreamer.Pad, error) {
	return v.gstVideobox.GetStaticPad("src")
}

func (v *rtcVideo) SetSize(width int64, height int64) {
	caps, _ := gstreamer.NewCapsFromString(fmt.Sprintf("%s,width=%d,height=%d", v.getCodecCaps(), width, height))
	v.gstSrc.Set("caps", caps)
}

func (v *rtcVideo) getCodecCaps() string {
	switch v.codec {
	case CodecVP8:
		return "application/x-rtp, encoding-name=VP8-DRAFT-IETF-01"
	case CodecVP9:
	case CodecH264:
	default:
		return "application/x-rtp"
	}

	return "application/x-rtp"
}

func (v *rtcVideo) internalLink() error {
	res := v.gstSrc.Link(v.gstFilter)
	if !res {
		return fmt.Errorf("(RTCVideo) Failed to link src with filter")
	}

	res = v.gstFilter.Link(v.gstDepay)
	if !res {
		return fmt.Errorf("Failed to link filter with depay")
	}

	res = v.gstDepay.Link(v.gstVideobox)
	if !res {
		return fmt.Errorf("Failed to link depay with videobox")
	}

	return nil
}

func (v *rtcVideo) SetPipeline(pipeline gstreamer.Pipeline) error {
	pipeline.Add(v.gstSrc)
	pipeline.Add(v.gstFilter)
	pipeline.Add(v.gstDepay)
	pipeline.Add(v.gstVideobox)

	return v.internalLink()
}

func (v *rtcVideo) Push(buffer []byte) {
	if v.gstSrc != nil {
		v.gstSrc.Push(buffer)
	}
}
