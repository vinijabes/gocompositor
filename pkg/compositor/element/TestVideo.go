package element

import (
	"fmt"

	"github.com/vinijabes/gostreamer/pkg/gstreamer"
)

type VideoTest interface {
	Video
}

type videoTest struct {
	video
	videofilter gstreamer.Element
	queue       gstreamer.Element
}

func NewVideoTest(width int, height int) (VideoTest, error) {
	video := &videoTest{}
	videosrc, err := gstreamer.NewElement("videotestsrc", fmt.Sprintf("source_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	videofilter, err := gstreamer.NewElement("capsfilter", fmt.Sprintf("videofilter_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	queue, err := gstreamer.NewElement("queue", fmt.Sprintf("queue_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	videobox, err := gstreamer.NewElement("videobox", fmt.Sprintf("box_%d", videoIDGenerator))
	if err != nil {
		return nil, err
	}

	video.videosrc = videosrc
	video.videofilter = videofilter
	video.queue = queue
	video.videobox = videobox

	videoIDGenerator++
	video.SetSize(width, height)

	return video, nil
}

func (v *videoTest) SetPipeline(pipeline gstreamer.Pipeline) error {
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
		!pipeline.Add(v.queue) ||
		!pipeline.Add(v.videobox) ||
		!v.videosrc.Link(v.videofilter) ||
		!v.videofilter.Link(v.queue) ||
		!v.queue.Link(v.videobox) {
		return ErrVideoSetPipeline
	}

	v.pipeline = pipeline

	return nil
}

func (v *videoTest) SetSize(width int, height int) {
	caps, _ := gstreamer.NewCapsFromString(fmt.Sprintf("%s,width=%d,height=%d", v.getCapsProps(), width, height))
	v.videofilter.Set("caps", caps)
}
