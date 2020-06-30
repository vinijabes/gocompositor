package compositor

import (
	"errors"
	"fmt"
	"time"

	"github.com/vinijabes/gocompositor/pkg/compositor/element"
	gstreamer "github.com/vinijabes/gostreamer/pkg/gstreamer"
)

//Compositor ...
type Compositor struct {
	pipeline gstreamer.Pipeline
	mixer    *Mixer
	layout   *Layout
	videos   element.Videos
}

//Mixer ...
type Mixer struct {
	gstMixer        gstreamer.Element
	gstOutputFilter gstreamer.Element
	gstPadTemplate  gstreamer.PadTemplate
}

var pipelineIDGenerator = 0

var ErrCreateCompositor = errors.New("Failed to create compositor")

func printBusMessages(bus gstreamer.Bus) {
	for {
		for bus.HavePending() {
			message, err := bus.Pop()

			if err != nil {
				fmt.Println(err)
			} else if message.GetType() == gstreamer.MessageError {
				fmt.Println(message.GetStructure())
			}
		}
		time.Sleep(1 * time.Second)
	}
}

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

	if !pipeline.Add(mixer.gstMixer) || !pipeline.Add(mixer.gstOutputFilter) {
		return nil, ErrCreateCompositor
	}

	bus, _ := pipeline.GetBus()
	go printBusMessages(bus)

	mixer.gstMixer.Link(mixer.gstOutputFilter)

	pipelineIDGenerator++
	return compositor, nil
}

func newMixer(id int) (*Mixer, error) {
	videomixer, err := gstreamer.NewElement("compositor", fmt.Sprintf("videomixer_%d", id))
	if err != nil {
		return nil, err
	}
	videomixer.Set("background", 1)

	padTemplate, err := videomixer.GetPadTemplate("sink_%u")
	if err != nil {
		return nil, err
	}

	capsfilter, err := gstreamer.NewElement("capsfilter", fmt.Sprintf("filter_%d", id))
	if err != nil {
		return nil, err
	}

	caps, err := gstreamer.NewCapsFromString(fmt.Sprintf("video/x-raw,width=%d,height=%d", 1280, 720))
	if err != nil {
		return nil, err
	}
	capsfilter.Set("caps", caps)

	mixer := &Mixer{
		gstMixer:        videomixer,
		gstPadTemplate:  padTemplate,
		gstOutputFilter: capsfilter,
	}

	return mixer, nil
}

//AddVideo add new video
func (c *Compositor) AddVideo(v element.Video) error {
	pipeline := c.pipeline

	err := v.SetPipeline(pipeline)
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
func (c *Compositor) Add(e gstreamer.Element) {
	c.pipeline.Add(e)
}

//Start ...
func (c *Compositor) Start() {
	c.pipeline.SetState(gstreamer.GstStatePlaying)
}

//Stop ...
func (c *Compositor) Stop() {
	c.pipeline.SetState(gstreamer.GstStateNull)
}

//Pause ...
func (c *Compositor) Pause() {
	c.pipeline.SetState(gstreamer.GstStatePaused)
}

//SetLayout ...
func (c *Compositor) SetLayout(l *Layout) {
	c.layout = l

	l.ApplyLayout(c.videos)
}

//LinkVideoSink ...
func (c *Compositor) LinkVideoSink(e gstreamer.Element) {
	c.mixer.gstOutputFilter.Link(e)
}

func (m *Mixer) link(v element.Video) error {
	sink, err := m.gstMixer.RequestPad(m.gstPadTemplate, nil, nil)
	if err != nil {
		return err
	}

	sink.Set("alpha", 1.0)
	result, err := v.LinkSinkPad(sink)

	if err != nil {
		return err
	}

	if result != gstreamer.GstPadLinkOk {
		return errors.New("Failed to link sink with video element")
	}

	return nil
}
