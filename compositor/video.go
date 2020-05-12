package compositor

import (
	"fmt"

	gstreamer "github.com/vinijabes/gostreamer"
)

type box struct {
	gstSink *gstreamer.Pad
}

//Video ...
type Video struct {
	pipeline    *gstreamer.Pipeline
	gstFilter   *gstreamer.Element
	gstVideobox *gstreamer.Element
	gstSrc      *gstreamer.Element
	id          int

	box *box
}

//Videos ...
type Videos []*Video

var videoIDGenerator = 0

//NewRawVideo ...
func NewRawVideo() *Video {
	return &Video{}
}

//NewTestVideo ...
func NewTestVideo(width int, height int, pattern int) (*Video, error) {
	video, err := NewVideo(width, height, "videotestsrc")
	if err != nil {
		return nil, err
	}

	video.gstSrc.SetInt("pattern", int64(pattern))

	return video, nil
}

//NewVideo ...
func NewVideo(width int, height int, srcPlugin string) (*Video, error) {
	src, err := gstreamer.NewElement(srcPlugin, fmt.Sprintf("source_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	filter, err := gstreamer.NewElement("capsfilter", fmt.Sprintf("filter_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	videobox, err := gstreamer.NewElement("videobox", fmt.Sprintf("box_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	video := &Video{
		gstSrc:      src,
		gstFilter:   filter,
		gstVideobox: videobox,
		id:          videoIDGenerator,
	}

	video.SetSize(int64(width), int64(height))

	videoIDGenerator++

	return video, nil
}

//LinkSrc link external element with src element
func (v *Video) LinkSrc(e *gstreamer.Element) error {
	return e.Link(v.gstSrc)
}

func (v *Video) internalLink() error {
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
func (v *Video) SetPos(x int64, y int64) {
	v.box.gstSink.SetInt("xpos", x)
	v.box.gstSink.SetInt("ypos", y)
}

//SetSize ...
func (v *Video) SetSize(width int64, height int64) {
	v.gstFilter.SetCapsFromString(fmt.Sprintf("video/x-raw,width=%d,height=%d", width, height))
}

//SetBorder ...
func (v *Video) SetBorder(border string, value int64) {
	v.gstVideobox.SetInt(border, value)
}
