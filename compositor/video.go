package compositor

import (
	"fmt"

	gstreamer "github.com/vinijabes/gostreamer"
)

type box struct {
	gstSink *gstreamer.Pad
}

type Video interface {
	SetPos(x int64, y int64)
	SetSize(width int64, height int64)
	SetBorder(border string, value int64)
	SetPipeline(pipeline *gstreamer.Pipeline) error
	SetBox(b *box)

	GetSrcPad() (*gstreamer.Pad, error)
}

type RTCVideo interface {
	Video

	Push(buffer []byte)
}

//Video ...
type video struct {
	pipeline    *gstreamer.Pipeline
	gstFilter   *gstreamer.Element
	gstVideobox *gstreamer.Element
	gstSrc      *gstreamer.Element
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
	gstDepay *gstreamer.Element
	codec    Codec
}

//Videos ...
type Videos []Video

var videoIDGenerator = 0

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
	raw.gstSrc.SetInt("pattern", int64(pattern))

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
func NewVideoFromElement(src *gstreamer.Element, width int, height int) (Video, error) {
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
func NewRTCVideo(width int, height int) (RTCVideo, error) {
	src, err := gstreamer.NewElement("appsrc", fmt.Sprintf("source_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}
	src.Set("format", "time")
	src.SetInt("is-live", int64(1))
	src.SetInt("do-timestamp", int64(1))

	filter, err := gstreamer.NewElement("capsfilter", fmt.Sprintf("filter_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	depay, err := gstreamer.NewElement(getDepayFromCodec(CodecVP8), fmt.Sprintf("depay_%d", videoIDGenerator))
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
		gstDepay: depay,
	}

	return v, nil
}

//LinkSrc link external element with src element
func (v *video) LinkSrc(e *gstreamer.Element) error {
	return e.Link(v.gstSrc)
}

func (v *video) internalLink() error {
	err := v.gstSrc.Link(v.gstFilter)
	if err != nil {
		return err
	}

	err = v.gstFilter.Link(v.gstVideobox)
	if err != nil {
		return err
	}

	return nil
}

//SetPos ...
func (v *video) SetPos(x int64, y int64) {
	v.box.gstSink.SetInt("xpos", x)
	v.box.gstSink.SetInt("ypos", y)
}

//SetSize ...
func (v *video) SetSize(width int64, height int64) {
	v.gstFilter.SetCapsFromString(fmt.Sprintf("%s,width=%d,height=%d", v.mediaType, width, height))
}

//SetBorder ...
func (v *video) SetBorder(border string, value int64) {
	v.gstVideobox.SetInt(border, value)
}

func (v *video) SetPipeline(pipeline *gstreamer.Pipeline) error {
	pipeline.Add(v.gstSrc)
	pipeline.Add(v.gstFilter)
	pipeline.Add(v.gstVideobox)

	return v.internalLink()
}

func (v *video) SetBox(b *box) {
	v.box = b
}

func (v *video) GetSrcPad() (*gstreamer.Pad, error) {
	return v.gstVideobox.GetStaticPad("src")
}

func (v *rtcVideo) SetSize(width int64, height int64) {
	v.gstFilter.SetCapsFromString(fmt.Sprintf("%s,width=%d,height=%d", v.getCodecCaps(), width, height))
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
	err := v.gstSrc.Link(v.gstFilter)
	if err != nil {
		return err
	}

	err = v.gstFilter.Link(v.gstDepay)
	if err != nil {
		return err
	}

	err = v.gstDepay.Link(v.gstVideobox)
	if err != nil {
		return err
	}

	return nil
}

func (v *rtcVideo) SetPipeline(pipeline *gstreamer.Pipeline) error {
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
