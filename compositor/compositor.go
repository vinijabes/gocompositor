package compositor

import (
	"errors"
	"fmt"

	gstreamer "github.com/vinijabes/gostreamer"
)

//Compositor ...
type Compositor struct {
	pipeline *gstreamer.Pipeline
	mixer    *Mixer
	layout   *Layout
	videos   Videos
}

//Mixer ...
type Mixer struct {
	gstMixer        *gstreamer.Element
	gstOutputFilter *gstreamer.Element
	gstPadTemplate  *gstreamer.PadTemplate
}

var pipelineIDGenerator = 0

//NewCompositor ...
func NewCompositor() (*Compositor, error) {
	pipeline, err := gstreamer.NewPipeline(fmt.Sprintf("compositor_%d", pipelineIDGenerator))
	if err != nil {
		return nil, err
	}

	mixer, err := newMixer(pipelineIDGenerator)
	if err != nil {
		return nil, err
	}

	compositor := &Compositor{
		pipeline: pipeline,
		mixer:    mixer,
	}

	pipeline.Add(mixer.gstMixer)
	pipeline.Add(mixer.gstOutputFilter)

	mixer.gstMixer.Link(mixer.gstOutputFilter)

	pipelineIDGenerator++
	return compositor, nil
}

func newMixer(id int) (*Mixer, error) {
	videomixer, err := gstreamer.NewElement("videomixer", fmt.Sprintf("videomixer_%d", id))
	if err != nil {
		return nil, err
	}
	videomixer.SetInt("background", 1)

	padTemplate, err := videomixer.GetPadTemplate("sink_%u")
	if err != nil {
		return nil, err
	}

	capsfilter, err := gstreamer.NewElement("capsfilter", fmt.Sprintf("filter_%d", id))
	if err != nil {
		return nil, err
	}

	capsfilter.SetCapsFromString(fmt.Sprintf("video/x-raw,width=%d,height=%d", 1280, 720))

	mixer := &Mixer{
		gstMixer:        videomixer,
		gstPadTemplate:  padTemplate,
		gstOutputFilter: capsfilter,
	}

	return mixer, nil
}

//AddVideo add new video
func (c *Compositor) AddVideo(v *Video) error {
	pipeline := c.pipeline
	v.pipeline = pipeline

	pipeline.Add(v.gstSrc)
	pipeline.Add(v.gstFilter)
	pipeline.Add(v.gstVideobox)

	err := v.internalLink()
	if err != nil {
		return err
	}

	err = c.mixer.link(v)
	if err != nil {
		return err
	}

	c.videos = append(c.videos, v)

	if c.layout != nil {
		c.layout.ApplyLayout(c.videos)
	}

	return nil
}

//Add ...
func (c *Compositor) Add(e *gstreamer.Element) {
	c.pipeline.Add(e)
}

//Start ...
func (c *Compositor) Start() {
	c.pipeline.Start()
}

//Stop ...
func (c *Compositor) Stop() {
	c.pipeline.Stop()
}

//Pause ...
func (c *Compositor) Pause() {
	c.pipeline.Pause()
}

//SetLayout ...
func (c *Compositor) SetLayout(l *Layout) {
	c.layout = l

	l.ApplyLayout(c.videos)
}

//LinkVideoSink ...
func (c *Compositor) LinkVideoSink(e *gstreamer.Element) {
	c.mixer.gstOutputFilter.Link(e)
}

func (m *Mixer) link(v *Video) error {
	sink, err := m.gstMixer.RequestPad(m.gstPadTemplate)
	if err != nil {
		return err
	}

	srcPad, err := v.gstVideobox.GetStaticPad("src")
	if err != nil {
		return err
	}

	if srcPad.Link(sink) == 0 {
		return errors.New("failed to link src pad with videomixer")
	}

	sink.SetFloat("alpha", 1.0)

	v.box = &box{gstSink: sink}

	return nil
}
