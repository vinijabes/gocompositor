package gstreamer

/*
#cgo pkg-config: gstreamer-1.0 gstreamer-base-1.0 gstreamer-app-1.0 gstreamer-plugins-base-1.0 gstreamer-video-1.0 gstreamer-audio-1.0 gstreamer-plugins-bad-1.0
#include "gstreamer.h"
*/
import "C"
import (
	"unsafe"
)

//Pipeline ...
type Pipeline struct {
	pipeline *C.GstPipeline
	id       int
}

func init() {
	C.gstreamer_init()
}

//NewPipeline ...
func NewPipeline(name string) (*Pipeline, error) {
	pipelineStrUnsafe := C.CString(name)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))
}
